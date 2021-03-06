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

func commUserIsNotMember(command string, ch string, username string, client net.Conn) {
	response := fmt.Sprintf("%s Failed: %s is not a member of %s\n", command, username, ch)
	communicate(response, client)
}

func commUserLeftChannel(command string, ch string, username string, client net.Conn) {
	response := fmt.Sprintf("%s Successful: %s was removed from %s\n", command, username, ch)
	communicate(response, client)
}

func commDestinationUserNotRegistered(command string, username string, client net.Conn) {
	response := fmt.Sprintf("%s Failed: %s is not a registered user\n", command, username)
	communicate(response, client)
}

func commDirectMessageDelivered(command string, username string, client net.Conn) {
	response := fmt.Sprintf("%s Successful: direct message delivered to %s\n", command, username)
	communicate(response, client)
}

func commFilenameAlreadyUsed(command string, filename string, client net.Conn) {
	response := fmt.Sprintf("%s Failed: file %s already exists, use another name\n", command, filename)
	communicate(response, client)
}

func commChannelWithNoFiles(command string, ch string, client net.Conn) {
	response := fmt.Sprintf("%s Successful: channel %s has no files\n", command, ch)
	communicate(response, client)
}

func commFileSaved(command string, ch string, filename string, client net.Conn) {
	response := fmt.Sprintf("%s Successful: %s saved in %s\n", command, filename, ch)
	communicate(response, client)
}

func commFileDoesntExist(command string, ch string, filename string, client net.Conn) {
	response := fmt.Sprintf("%s Failed: file %s doesn't exist in channel %s\n", command, filename, ch)
	communicate(response, client)
}

func commSendingFile(command string, filename string, client net.Conn) {
	response := fmt.Sprintf("%s Successful: sending %s file", command, filename)
	communicate(response, client)
}

func commNoChannelsCreated(command string, client net.Conn) {
	response := fmt.Sprintf("%s Successful: There are no channels created\n", command)
	communicate(response, client)
}
