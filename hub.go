package main

import "fmt"

type hub struct {
	channels        map[string]*channel
	clients         map[string]*client
	commands        chan command
	deregistrations chan *client
	registrations   chan *client
}

func (h *hub) run() {
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
				fmt.Errorf("Hub Switch: Cannot process %s", cmd.id)
			}
		}
	}
}

func (h *hub) register(c *client) {
	if _, exists := h.clients[c.username]; exists {
		c.username = ""
		c.conn.Write([]byte("ERR username taken\n"))
	} else {
		h.clients[c.username] = c
		c.conn.Write([]byte("OK\n"))
	}
}

func (h *hub) deregister(c *client) {
	if _, exists := h.clients[c.username]; exists {
		delete(h.clients, c.username)
		for _, channel := range h.channels {
			delete(channel.clients, c)
		}
	}
}

func (h *hub) joinChannel(u string, c string) {
	if client, ok := h.clients[u]; ok {
		if channel, ok := h.channels[c]; ok {
			// Channel exists, join
			channel.clients[client] = true
		} else {
			// Channel doesn't exists, create and join
			h.channels[c] = newChannel(c)
			h.channels[c].clients[client] = true
		}
	}
}
