package main

type ID int

const (
	REG ID = iota
	JOIN
	LEAVE
	MSG
	CHNS
	USRS
	SEND
	GET
)

type command struct {
	id        ID
	recipient string
	sender    *client
	body      []byte
}
