package main

import (
	"bytes"
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
			case FILES:
				h.listFiles(cmd.sender, cmd.recipient)
			case SEND:
				h.sendFile(cmd.sender, cmd.recipient, cmd.header, cmd.body)
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
	if h.userRegistered(cl.username) {
		commUsernameTaken(cl.username, cl.conn)
		cl.username = ""
	} else {
		h.clients[cl.username] = cl
		response = fmt.Sprintf("REG Successful: registered as %s \n", cl.username)
		communicate(response, cl.conn)
	}
}

func (h *hub) deregister(cl *client) {
	if h.userRegistered(cl.username) {
		delete(h.clients, cl.username)
		for _, channel := range h.channels {
			delete(channel.clients, cl)
		}
		log.Printf("DREG Executed: connection lost with %s \n", cl.username)
	}
}

func (h *hub) joinChannel(cl *client, ch string) {
	var response string

	if !h.userRegistered(cl.username) {
		response = "JOIN Failed: user isn't registered\n"
		communicate(response, cl.conn)
		return
	}
	client := h.clients[cl.username]

	if !h.channelExists(ch) {
		h.channels[ch] = newChannel(ch)
		h.channels[ch].clients[client] = true
		response = fmt.Sprintf("JOIN Successful: channel %s was created and user %s has joined it\n", ch, cl.username)
		communicate(response, cl.conn)
		return
	}
	channel := h.channels[ch]

	channel.clients[client] = true
	response = fmt.Sprintf("JOIN Successful: %s was added to %s\n", cl.username, ch)
	communicate(response, cl.conn)
}

func (h *hub) leaveChannel(cl *client, ch string) {
	var response string

	if !h.userRegistered(cl.username) {
		response = "LEAVE Failed: user isn't registered\n"
		communicate(response, cl.conn)
		return
	}
	client := h.clients[cl.username]

	if !h.channelExists(ch) {
		response = fmt.Sprintf("LEAVE Failed: channel %s doesn't exist\n", ch)
		communicate(response, cl.conn)
		return
	}
	channel := h.channels[ch]

	if !h.userIsMember(channel, client) {
		response = fmt.Sprintf("LEAVE Failed: %s is not a member of %s\n", cl.username, ch)
		communicate(response, cl.conn)
		return
	}

	delete(channel.clients, client)
	response = fmt.Sprintf("LEAVE Successful: %s was removed from %s\n", cl.username, ch)
	communicate(response, cl.conn)

}

func (h *hub) message(cl *client, recipient string, m []byte) {
	var response string

	if !h.userRegistered(cl.username) {
		response = "MSG Failed: user isn't registered\n"
		communicate(response, cl.conn)
		return
	}
	sender := h.clients[cl.username]

	switch recipient[0] {
	case '#':
		if !h.channelExists(recipient) {
			response = fmt.Sprintf("MSG Failed: channel %s doesn't exist\n", recipient)
			communicate(response, cl.conn)
			return
		}
		channel := h.channels[recipient]

		if !h.userIsMember(channel, sender) {
			response = fmt.Sprintf("MSG Failed: %s is not a member of %s\n", cl.username, recipient)
			communicate(response, cl.conn)
			return
		}
		channel.broadcast(sender.username, m)
		log.Printf("MSG Successful: %s sent a message to %s\n", cl.username, recipient)
	case '@':
		if !h.userRegistered(recipient) {
			response = fmt.Sprintf("MSG Failed: %s is not a registered user\n", recipient)
			communicate(response, cl.conn)
			return
		}
		user := h.clients[recipient]

		msg := append([]byte(cl.username), ": "...)
		msg = append(msg, m...)
		msg = append(msg, "\n"...)

		user.conn.Write(msg)
		response = fmt.Sprintf("MSG Successful: message delivered to %s\n", user.username)
		communicate(response, cl.conn)

	}
}

func (h *hub) listFiles(cl *client, ch string) {
	var response string

	if !h.userRegistered(cl.username) {
		response = "FILES Failed: user isn't registered\n"
		communicate(response, cl.conn)
		return
	}
	sender := h.clients[cl.username]

	if !h.channelExists(ch) {
		response = fmt.Sprintf("FILES Failed: channel %s doesn't exist\n", ch)
		communicate(response, cl.conn)
		return
	}
	channel := h.channels[ch]

	if !h.userIsMember(channel, sender) {
		response = fmt.Sprintf("FILES Failed: %s is not a member of %s\n", cl.username, ch)
		communicate(response, cl.conn)
		return
	}

	var files []string

	for file := range channel.files {
		files = append(files, file)
	}

	enum := strings.Join(files, "\n")
	list := "Channel files ->>\n" + enum

	cl.conn.Write([]byte(list + "\n"))
	log.Printf("FILES Successful: list delivered to %s\n", cl.username)
}

func (h *hub) sendFile(cl *client, ch string, filename []byte, file []byte) {
	var response string

	if !h.userRegistered(cl.username) {
		response = "SEND Failed: user isn't registered\n"
		communicate(response, cl.conn)
		return
	}
	sender := h.clients[cl.username]

	if !h.channelExists(ch) {
		response = fmt.Sprintf("SEND Failed: channel %s doesn't exist\n", ch)
		communicate(response, cl.conn)
		return
	}
	channel := h.channels[ch]

	if !h.userIsMember(channel, sender) {
		response = fmt.Sprintf("SEND Failed: %s is not a member of %s\n", cl.username, ch)
		communicate(response, cl.conn)
		return
	}

	fn := string(filename)

	if h.fileExists(channel, fn) {
		response = fmt.Sprintf("SEND Failed: file %s already exists, use another name\n", fn)
		communicate(response, cl.conn)
		return
	}

	fileAddress := newFile(fn, file)
	h.channels[ch].files[fn] = fileAddress

	response = fmt.Sprintf("SEND Successful: %s saved in %s\n", fn, ch)
	communicate(response, cl.conn)

	channel.broadcast(cl.username, []byte(fmt.Sprintf("just saved %s file", fn)))
}

func (h *hub) getFile(cl *client, ch string, filename []byte) {
	var response string

	if !h.userRegistered(cl.username) {
		response = "GET Failed: user isn't registered\n"
		communicate(response, cl.conn)
		return
	}
	sender := h.clients[cl.username]

	if !h.channelExists(ch) {
		response = fmt.Sprintf("GET Failed: channel %s doesn't exist\n", ch)
		communicate(response, cl.conn)
		return
	}
	channel := h.channels[ch]

	if !h.userIsMember(channel, sender) {
		response = fmt.Sprintf("GET Failed: %s is not a member of %s\n", cl.username, ch)
		communicate(response, cl.conn)
		return
	}

	fn := string(filename)
	if !h.fileExists(channel, fn) {
		response = fmt.Sprintf("GET Failed: file %s doesn't exist\n", fn)
		communicate(response, cl.conn)
		return
	}

	var payload []byte
	formattedBody := replaceReturns(channel.files[fn].body)

	payload = append(payload, []byte("FILE ")...)
	payload = append(payload, filename...)
	payload = append(payload, []byte(" ")...)
	payload = append(payload, formattedBody...)

	response = fmt.Sprintf("GET Successful: sending %s file", fn)
	communicate(response, cl.conn)
	cl.conn.Write(payload)
}

func (h *hub) listUsers(cl *client) {
	var response string

	if !h.userRegistered(cl.username) {
		response = "USRS Failed: user isn't registered\n"
		communicate(response, cl.conn)
		return
	}
	client := h.clients[cl.username]

	var names []string

	for c := range h.clients {
		names = append(names, c)
	}

	enum := strings.Join(names, ", ")
	list := "->> Registered users: " + enum

	client.conn.Write([]byte(list + "\n"))
	log.Printf("USRS Successful: list delivered to %s\n", cl.username)
}

func (h *hub) listChannels(cl *client) {
	var response string

	if !h.userRegistered(cl.username) {
		response = "CHNS Failed: user isn't registered\n"
		communicate(response, cl.conn)
		return
	}
	client := h.clients[cl.username]

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
}

func replaceReturns(body []byte) []byte {
	splittedBody := bytes.Split(body, []byte("\n"))
	joinedBody := bytes.Join(splittedBody, []byte("//"))
	joinedBody = append(joinedBody, '\n')
	return joinedBody
}
