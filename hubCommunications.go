package main

import (
	"fmt"
	"log"
	"net"
)

func communicate(response string, connection net.Conn) {
	log.Print(response)
	connection.Write([]byte(string("->> " + response + "\n")))
}

func commUsernameTaken(username string, client net.Conn) {
	response := fmt.Sprintf("REG Denied: %s was already taken\n", username)
	communicate(response, client)
}

func commRegisterSuccess(username string, client net.Conn) {
	response := fmt.Sprintf("REG Successful: registered as %s \n", username)
	communicate(response, client)
}
