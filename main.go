package main

import (
	"log"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Printf("%v", err)
	}

	logs := make(chan string)
	hub := newHub(logs)
	go hub.run()
	log.Println(<-hub.hubLogs)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("%v", err)
		}

		c := newClient(
			conn,
			hub.commands,
			hub.registrations,
			hub.deregistrations,
		)
		go c.read()

		log.Printf("New connection received and connected to the hub service channels...")
		conn.Write([]byte("Welcome to miniSlack! Please use REG to register along with @username (provide your username instead)\n"))
		log.Println(<-hub.hubLogs)
	}
}
