package main

import (
	"bytes"
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
			case FILES:
				h.listFiles(cmd.sender, cmd.recipient)
			case SEND:
				h.sendFile(cmd.sender, cmd.recipient, cmd.body)
			case GET:
				h.getFile(cmd.sender, cmd.recipient, cmd.body)
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

func (h *hub) listFiles(cl *client, ch string) {
	var response string
	if sender, ok := h.clients[cl.username]; ok {
		if channel, ok := h.channels[ch]; ok {
			if _, ok := h.channels[ch].clients[sender]; ok {
				var files []string

				for file := range channel.files {
					files = append(files, file)
				}

				enum := strings.Join(files, "\n")
				list := "Channel files ->>\n" + enum

				cl.conn.Write([]byte(list + "\n"))
				log.Printf("FILES Successful: list delivered to %s\n", cl.username)
			} else {
				response = fmt.Sprintf("FILES Failed: %s is not a member of %s\n", cl.username, ch)
				communicate(response, cl.conn)
			}
		} else {
			response = fmt.Sprintf("FILES Failed: channel %s doesn't exist\n", ch)
			communicate(response, cl.conn)
		}
	} else {
		response = "FILES Failed: user isn't registered\n"
		communicate(response, cl.conn)
	}
}

func (h *hub) sendFile(cl *client, ch string, file []byte) {
	var response string

	if _, ok := h.clients[cl.username]; !ok {
		response = "SEND Failed: user isn't registered\n"
		communicate(response, cl.conn)
		return
	}
	sender := h.clients[cl.username]

	if _, ok := h.channels[ch]; !ok {
		response = fmt.Sprintf("SEND Failed: channel %s doesn't exist\n", ch)
		communicate(response, cl.conn)
		return
	}
	channel := h.channels[ch]

	if _, ok := channel.clients[sender]; !ok {
		response = fmt.Sprintf("SEND Failed: %s is not a member of %s\n", cl.username, ch)
		communicate(response, cl.conn)
		return
	}

	filename := bytes.Split(file, []byte("\n"))[0]
	fn := string(filename)

	if _, ok := channel.files[fn]; ok {
		response = fmt.Sprintf("SEND Failed: file %s already exists, use another name\n", fn)
		communicate(response, cl.conn)
		return
	}

	body := bytes.TrimPrefix(file, filename)
	fileAddress := newFile(fn, body)
	h.channels[ch].files[fn] = fileAddress

	response = fmt.Sprintf("SEND Successful: %s saved in %s\n", fn, ch)
	communicate(response, cl.conn)

	channel.broadcast(cl.username, []byte(fmt.Sprintf("just saved %s file", fn)))
}

func (h *hub) getFile(cl *client, ch string, filename []byte) {
	var response string

	if _, ok := h.clients[cl.username]; !ok {
		response = "GET Failed: user isn't registered\n"
		communicate(response, cl.conn)
		return
	}
	sender := h.clients[cl.username]

	if _, ok := h.channels[ch]; !ok {
		response = fmt.Sprintf("GET Failed: channel %s doesn't exist\n", ch)
		communicate(response, cl.conn)
		return
	}
	channel := h.channels[ch]

	if _, ok := channel.clients[sender]; !ok {
		response = fmt.Sprintf("GET Failed: %s is not a member of %s\n", cl.username, ch)
		communicate(response, cl.conn)
		return
	}

	fn := string(filename)
	log.Println("fn:", fn)
	if _, ok := channel.files[fn]; !ok {
		response = fmt.Sprintf("GET Failed: file %s doesn't exist\n", fn)
		communicate(response, cl.conn)
		return
	}

	cl.conn.Write([]byte("File exists\n"))
}

func (h *hub) listUsers(cl *client) {
	var response string
	if client, ok := h.clients[cl.username]; ok {
		var names []string

		for c := range h.clients {
			names = append(names, c)
		}

		enum := strings.Join(names, ", ")
		list := "->> Registered users: " + enum

		client.conn.Write([]byte(list + "\n"))
		log.Printf("USRS Successful: list delivered to %s\n", cl.username)
	} else {
		response = "USRS Failed: user isn't registered\n"
		communicate(response, cl.conn)
	}
}

func (h *hub) listChannels(cl *client) {
	var response string
	if client, ok := h.clients[cl.username]; ok {
		var names []string

		if len(h.channels) == 0 {
			response = "CHNS Successful: There are no channels created\n"
			communicate(response, cl.conn)
		} else {
			for c := range h.channels {
				names = append(names, c)
			}
			enum := strings.Join(names, ", ")
			list := "->> Channels: " + enum
			client.conn.Write([]byte(list + "\n"))
			log.Printf("CHNS Successful: list delivered to %s", cl.username)
		}
	} else {
		response = "CHNS Failed: user isn't registered\n"
		communicate(response, cl.conn)
	}
}

func communicate(response string, connection net.Conn) {
	log.Print(response)
	connection.Write([]byte(string("->> " + response)))
}
