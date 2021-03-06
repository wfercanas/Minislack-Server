package main

type ID int

const (
	REG ID = iota
	JOIN
	LEAVE
	MSG
	CHNS
	USRS
	FILES
	SEND
	GET
)

type command struct {
	id        ID
	recipient string
	sender    *client
	header    []byte
	body      []byte
}
