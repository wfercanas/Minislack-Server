package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
)

var DELIMITER = []byte(`\r\n`)

type client struct {
	conn       net.Conn
	outbound   chan<- command
	register   chan<- *client
	deregister chan<- *client
	username   string
}

func (c *client) read() error {
	for {
		msg, err := bufio.NewReader(c.conn).ReadBytes('\n')
		if err == io.EOF {
			// Coinnection closed, deregister client
			c.deregister <- c
			return nil
		}
		if err != nil {
			return err
		}

		c.handle(msg)
	}
}

func (c *client) handle(message []byte) {
	cmd := bytes.ToUpper(bytes.TrimSpace(bytes.Split(message, []byte(" "))[0]))
	args := bytes.TrimSpace(bytes.TrimPrefix(message, cmd))

	switch string(cmd) {
	case "REG":
		if err := c.reg(args); err != nil {
			c.err(err)
		}
	case "JOIN":
		if err := c.join(args); err != nil {
			c.err(err)
		}
	case "LEAVE":
		if err := c.leave(args); err != nil {
			c.err(err)
		}
	case "MSG":
		if err := c.msg(args); err != nil {
			c.err(err)
		}
	case "CHNS":
		c.chns()
	case "USRS":
		c.usrs()
	default:
		c.err(fmt.Errorf("Unknow command %s", cmd))
	}
}

func (c *client) reg(args []byte) error {
	u := bytes.TrimSpace(args)
	if u[0] != '@' {
		return fmt.Errorf("Username must begin with @")
	}
	if len(u) == 0 {
		return fmt.Errorf("Username cannot be blank")
	}

	c.username = string(u)
	c.register <- c

	return nil
}

func (c *client) msg(args []byte) error {
	args = bytes.TrimSpace(args)
	if args[0] != '#' && args[0] != '@' {
		return fmt.Errorf("Recipient must be a channel ('#name') or a user (''@user)")
	}

	recipient := bytes.Split(args, []byte(" "))[0]
	if len(recipient) == 0 {
		return fmt.Errorf("Recipient must have a name")
	}

	args = bytes.TrimSpace(bytes.TrimPrefix(args, recipient))
	l := bytes.Split(args, DELIMITER)[0]
	length, err := strconv.Atoi(string(l))
	if err != nil {
		return fmt.Errorf("Body length must be present")
	}
	if length == 0 {
		return fmt.Errorf("Body length must be at least 1")
	}

	padding := len(l) + len(DELIMITER) // Size of the body length + delimiter
	body := args[padding : padding+length]

	c.outbound <- command{
		recipient: string(recipient[1:]),
		sender:    c.username,
		body:      body,
		id:        MSG,
	}

	return nil
}

func (c *client) err(e error) {
	c.conn.Write([]byte("ERR " + e.Error() + "\n"))
}
