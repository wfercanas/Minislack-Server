package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type hub struct {
	channels        map[string]*channel
	clients         map[string]*client
	commands        chan command
	deregistrations chan *client
	registrations   chan *client
}

func newHub() *hub {
	return &hub{
		channels:        make(map[string]*channel),
		clients:         make(map[string]*client),
		commands:        make(chan command),
		registrations:   make(chan *client),
		deregistrations: make(chan *client),
	}
}

func communicate(response string, connection net.Conn) {
	log.Print(response)
	connection.Write([]byte(string("->> " + response)))
}

func (h *hub) run() {
	log.Println("Hub up and running...")
	for {
		select {
		case client := <-h.registrations:
			h.register(client)
		case client := <-h.deregistrations:
			h.deregister(client)
		case cmd := <-h.commands:
			switch cmd.id {
			case JOIN:
				h.joinChannel(cmd.sender, cmd.recipient)
			case LEAVE:
				h.leaveChannel(cmd.sender, cmd.recipient)
			case MSG:
				h.message(cmd.sender, cmd.recipient, cmd.body)
			case USRS:
				h.listUsers(cmd.sender)
			case CHNS:
				h.listChannels(cmd.sender)
			default:
				// fmt.Errorf("Hub Switch: Cannot process %s", cmd.id)
			}
		}
	}
}

func (h *hub) register(cl *client) {
	var response string
	if _, exists := h.clients[cl.username]; exists {
		response = fmt.Sprintf("REG Denied: %s was already taken\n", cl.username)
		communicate(response, cl.conn)
		cl.username = ""
	} else {
		h.clients[cl.username] = cl
		response = fmt.Sprintf("REG Successful: registered as %s \n", cl.username)
		communicate(response, cl.conn)
	}
}

func (h *hub) deregister(cl *client) {
	if _, exists := h.clients[cl.username]; exists {
		delete(h.clients, cl.username)
		for _, channel := range h.channels {
			delete(channel.clients, cl)
		}
	}
	log.Printf("DREG Executed: connection lost with %s \n", cl.username)
}

func (h *hub) joinChannel(cl *client, ch string) {
	var response string
	if client, ok := h.clients[cl.username]; ok {
		if channel, ok := h.channels[ch]; ok {
			channel.clients[client] = true
			response = fmt.Sprintf("JOIN Successful: %s was added to %s\n", cl.username, ch)
			communicate(response, cl.conn)
		} else {
			h.channels[ch] = newChannel(ch)
			h.channels[ch].clients[client] = true
			response = fmt.Sprintf("JOIN Successful: channel %s was created and user %s has joined it\n", ch, cl.username)
			communicate(response, cl.conn)
		}
	} else {
		response = "JOIN Failed: user isn't registered\n"
		communicate(response, cl.conn)
	}
}

func (h *hub) leaveChannel(cl *client, ch string) {
	var response string
	if client, ok := h.clients[cl.username]; ok {
		if channel, ok := h.channels[ch]; ok {
			delete(channel.clients, client)
			response = fmt.Sprintf("LEAVE Successful: %s was removed from %s\n", cl.username, ch)
			communicate(response, cl.conn)
		} else {
			response = fmt.Sprintf("LEAVE Failed: %s doesn't exist\n", ch)
			communicate(response, cl.conn)
		}
	} else {
		response = "LEAVE Failed: user isn't registered\n"
		communicate(response, cl.conn)
	}
}

func (h *hub) message(cl *client, r string, m []byte) {
	var response string
	if sender, ok := h.clients[cl.username]; ok {
		switch r[0] {
		case '#':
			if channel, ok := h.channels[r]; ok {
				if _, ok := channel.clients[sender]; ok {
					channel.broadcast(sender.username, m)
					log.Printf("MSG Successful: %s sent a message to %s\n", cl.username, r)
				} else {
					response = fmt.Sprintf("MSG Failed: %s is not a member of %s\n", cl.username, r)
					communicate(response, cl.conn)
				}
			} else {
				response = fmt.Sprintf("MSG Failed: %s doesn't exist\n", r)
				communicate(response, cl.conn)
			}
		case '@':
			if user, ok := h.clients[r]; ok {
				msg := append([]byte(cl.username), ": "...)
				msg = append(msg, m...)
				msg = append(msg, "\n"...)
				user.conn.Write(msg)
				response = fmt.Sprintf("MSG Successful: message delivered to %s\n", user.username)
				communicate(response, cl.conn)
			} else {
				response = fmt.Sprintf("MSG Failed: %s is not a registered user\n", r)
				communicate(response, cl.conn)
			}
		}
	} else {
		response = "MSG Failed: user isn't registered\n"
		communicate(response, cl.conn)
	}
}

func (h *hub) listUsers(cl *client) {
	if client, ok := h.clients[cl.username]; ok {
		var names []string

		for c := range h.clients {
			names = append(names, "@"+c+" ")
		}

		resp := strings.Join(names, ", ")
		client.conn.Write([]byte(resp + "\n"))
	}
}

func (h *hub) listChannels(cl *client) {
	if client, ok := h.clients[cl.username]; ok {
		var names []string

		if len(h.channels) == 0 {
			client.conn.Write([]byte("ERR no channels found\n"))
		}

		for c := range h.channels {
			names = append(names, "#"+c+" ")
		}

		resp := strings.Join(names, ", ")
		client.conn.Write([]byte(resp + "\n"))
	}
}
