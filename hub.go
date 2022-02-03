package main

import (
	"fmt"
	"log"
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

func (h *hub) register(c *client) {
	var response string
	if _, exists := h.clients[c.username]; exists {
		response = fmt.Sprintf("REG Denied: %s was already taken\n", c.username)
		c.conn.Write([]byte(response))
		log.Print(response)
		c.username = ""
	} else {
		h.clients[c.username] = c
		response = fmt.Sprintf("REG Successful: registered as %s \n", c.username)
		c.conn.Write([]byte(response))
		log.Print(response)
	}
}

func (h *hub) deregister(c *client) {
	if _, exists := h.clients[c.username]; exists {
		delete(h.clients, c.username)
		for _, channel := range h.channels {
			delete(channel.clients, c)
		}
	}
	log.Printf("DREG Executed: connection lost with %s \n", c.username)
}

func (h *hub) joinChannel(cl *client, ch string) {
	var response string
	if client, ok := h.clients[cl.username]; ok {
		if channel, ok := h.channels[ch]; ok {
			channel.clients[client] = true
			response = fmt.Sprintf("JOIN Successful: %s was added to %s\n", cl.username, ch)
			log.Print(response)
			cl.conn.Write([]byte(response))
		} else {
			h.channels[ch] = newChannel(ch)
			h.channels[ch].clients[client] = true
			response = fmt.Sprintf("JOIN Successful: channel %s was created and user %s has joined it\n", ch, cl.username)
			log.Print(response)
			cl.conn.Write([]byte(response))
		}
	} else {
		response = "JOIN Failed: user isn't registered\n"
		log.Print(response)
		cl.conn.Write([]byte(response))
	}
}

func (h *hub) leaveChannel(cl *client, ch string) {
	if client, ok := h.clients[cl.username]; ok {
		if channel, ok := h.channels[ch]; ok {
			delete(channel.clients, client)
		}
	}
}

func (h *hub) message(cl *client, r string, m []byte) {
	if sender, ok := h.clients[cl.username]; ok {
		switch r[0] {
		case '#':
			if channel, ok := h.channels[r]; ok {
				if _, ok := channel.clients[sender]; ok {
					channel.broadcast(sender.username, m)
				}
			}
		case '@':
			if user, ok := h.clients[r]; ok {
				user.conn.Write(append(m, '\n'))
			}
		}
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
