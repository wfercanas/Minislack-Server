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
	if h.userRegistered(cl.username) {
		commUsernameTaken(cl.username, cl.conn)
		cl.username = ""
	} else {
		h.clients[cl.username] = cl
		commRegisterSuccess(cl.username, cl.conn)
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
	if !h.userRegistered(cl.username) {
		commUserNotRegistered("JOIN", cl.conn)
		return
	}

	if !h.channelExists(ch) {
		h.channels[ch] = newChannel(ch)
		commChannelCreated("JOIN", ch, cl.conn)
	}
	channel := h.channels[ch]

	channel.clients[cl] = true
	commUserJoinedChannel("JOIN", ch, cl.username, cl.conn)
}

func (h *hub) leaveChannel(cl *client, ch string) {
	if !h.userRegistered(cl.username) {
		commUserNotRegistered("LEAVE", cl.conn)
		return
	}

	if !h.channelExists(ch) {
		commChannelDoesntExist("LEAVE", ch, cl.conn)
		return
	}
	channel := h.channels[ch]

	if !h.userIsMember(channel, cl) {
		commUserIsNotMember("LEAVE", ch, cl.username, cl.conn)
		return
	}

	delete(channel.clients, cl)
	commUserLeftChannel("LEAVE", ch, cl.username, cl.conn)
}

func (h *hub) message(cl *client, recipient string, m []byte) {
	if !h.userRegistered(cl.username) {
		commUserNotRegistered("MSG", cl.conn)
		return
	}

	switch recipient[0] {
	case '#':
		if !h.channelExists(recipient) {
			commChannelDoesntExist("MSG", recipient, cl.conn)
			return
		}
		channel := h.channels[recipient]

		if !h.userIsMember(channel, cl) {
			commUserIsNotMember("MSG", recipient, cl.username, cl.conn)
			return
		}

		channel.broadcast(cl.username, m)
		log.Printf("MSG Successful: %s sent a message to %s\n", cl.username, recipient)
	case '@':
		if !h.userRegistered(recipient) {
			commDestinationUserNotRegistered("MSG", recipient, cl.conn)
			return
		}
		user := h.clients[recipient]

		msg := append([]byte(cl.username), ": "...)
		msg = append(msg, m...)
		msg = append(msg, "\n"...)
		user.conn.Write(msg)

		commDirectMessageDelivered("MSG", user.username, cl.conn)
	}
}

func (h *hub) listFiles(cl *client, ch string) {
	if !h.userRegistered(cl.username) {
		commUserNotRegistered("FILES", cl.conn)
		return
	}

	if !h.channelExists(ch) {
		commChannelDoesntExist("FILES", ch, cl.conn)
		return
	}
	channel := h.channels[ch]

	if !h.userIsMember(channel, cl) {
		commUserIsNotMember("FILES", ch, cl.username, cl.conn)
		return
	}

	var files []string

	if len(channel.files) == 0 {
		commChannelWithNoFiles("FILES", ch, cl.conn)
	} else {
		for file := range channel.files {
			files = append(files, file)
		}

		enum := strings.Join(files, ", ")
		list := "Channel files ->>\n" + enum

		cl.conn.Write([]byte(list + "\n"))
		log.Printf("FILES Successful: list delivered to %s\n", cl.username)
	}
}

func (h *hub) sendFile(cl *client, ch string, filename []byte, file []byte) {
	if !h.userRegistered(cl.username) {
		commUserNotRegistered("SEND", cl.conn)
		return
	}

	if !h.channelExists(ch) {
		commChannelDoesntExist("SEND", ch, cl.conn)
		return
	}
	channel := h.channels[ch]

	if !h.userIsMember(channel, cl) {
		commUserIsNotMember("SEND", ch, cl.username, cl.conn)
		return
	}

	fn := string(filename)
	if h.fileExists(channel, fn) {
		commFilenameAlreadyUsed("SEND", fn, cl.conn)
		return
	}

	fileAddress := newFile(fn, file)
	h.channels[ch].files[fn] = fileAddress

	commFileSaved("SEND", ch, fn, cl.conn)
	channel.broadcast(cl.username, []byte(fmt.Sprintf("just saved %s file", fn)))
}

func (h *hub) getFile(cl *client, ch string, filename []byte) {
	if !h.userRegistered(cl.username) {
		commUserNotRegistered("GET", cl.conn)
		return
	}

	if !h.channelExists(ch) {
		commChannelDoesntExist("GET", ch, cl.conn)
		return
	}
	channel := h.channels[ch]

	if !h.userIsMember(channel, cl) {
		commUserIsNotMember("GET", ch, cl.username, cl.conn)
		return
	}

	fn := string(filename)
	if !h.fileExists(channel, fn) {
		commFileDoesntExist("GET", ch, fn, cl.conn)
		return
	}

	var payload []byte
	formattedBody := replaceReturns(channel.files[fn].body)

	payload = append(payload, []byte("FILE ")...)
	payload = append(payload, filename...)
	payload = append(payload, []byte(" ")...)
	payload = append(payload, formattedBody...)

	commSendingFile("GET", fn, cl.conn)
	cl.conn.Write(payload)
}

func (h *hub) listUsers(cl *client) {
	if !h.userRegistered(cl.username) {
		commUserNotRegistered("USRS", cl.conn)
		return
	}

	var names []string

	for c := range h.clients {
		names = append(names, c)
	}

	enum := strings.Join(names, ", ")
	list := "->> Registered users: " + enum

	cl.conn.Write([]byte(list + "\n"))
	log.Printf("USRS Successful: list delivered to %s\n", cl.username)
}

func (h *hub) listChannels(cl *client) {
	if !h.userRegistered(cl.username) {
		commUserNotRegistered("CHNS", cl.conn)
		return
	}

	var names []string

	if len(h.channels) == 0 {
		commNoChannelsCreated("CHNS", cl.conn)
	} else {
		for c := range h.channels {
			names = append(names, c)
		}
		enum := strings.Join(names, ", ")
		list := "->> Channels: " + enum
		cl.conn.Write([]byte(list + "\n"))
		log.Printf("CHNS Successful: list delivered to %s", cl.username)
	}
}
