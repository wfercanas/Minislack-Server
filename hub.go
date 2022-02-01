package main

type hub struct {
	channels        map[string]*channel
	clients         map[string]*client
	commands        chan command
	deregistrations chan *client
	registrations   chan *client
}
