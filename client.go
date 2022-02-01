package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
)

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
