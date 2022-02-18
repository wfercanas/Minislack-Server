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

func commUserNotRegistered(command string, client net.Conn) {
	response := fmt.Sprintf("%s Failed: user isn't registered\n", command)
	communicate(response, client)
}

func commChannelCreated(command string, ch string, client net.Conn) {
	response := fmt.Sprintf("%s Successful: channel %s was created", command, ch)
	communicate(response, client)
}

func commUserJoinedChannel(command string, ch string, username string, client net.Conn) {
	response := fmt.Sprintf("%s Successful: %s was added to %s\n", command, username, ch)
	communicate(response, client)
}

func commChannelDoesntExist(command string, ch string, client net.Conn) {
	response := fmt.Sprintf("%s Failed: channel %s doesn't exist\n", command, ch)
	communicate(response, client)
}
