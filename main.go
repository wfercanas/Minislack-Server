package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Printf("%v", err)
	}

	hub := newHub()
	go hub.run()
	fmt.Println("Server up and hub running...")

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
	}
}
