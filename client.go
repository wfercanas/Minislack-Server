package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
)

var MSG_DELIMITER = []byte(`\r\n`)
var FILE_DELIMITER = []byte(" >> ")
var BREAK_LINE_DELIMITER = []byte("//")

type client struct {
	conn       net.Conn
	outbound   chan<- command
	register   chan<- *client
	deregister chan<- *client
	username   string
}

func newClient(conn net.Conn, o chan<- command, r chan<- *client, d chan<- *client) *client {
	return &client{
		conn:       conn,
		outbound:   o,
		register:   r,
		deregister: d,
	}
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
	cmd := bytes.TrimSpace(bytes.Split(message, []byte(" "))[0])
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
	case "FILES":
		if err := c.files(args); err != nil {
			c.err(err)
		}
	case "SEND":
		if err := c.send(args); err != nil {
			c.err(err)
		}
	case "GET":
		if err := c.get(args); err != nil {
			c.err(err)
		}
	case "CHNS":
		c.chns()
	case "USRS":
		c.usrs()
	default:
		c.err(fmt.Errorf("->> ERR: Unknown command %s", cmd))
	}
}

func (c *client) reg(args []byte) error {
	u := bytes.TrimSpace(args)
	if u[0] != '@' {
		return fmt.Errorf("->> ERR: Username must begin with @")
	}
	if len(u) == 0 {
		return fmt.Errorf("->> ERR: Username cannot be blank")
	}

	c.username = string(u)
	c.register <- c

	return nil
}

func (c *client) join(args []byte) error {
	channelID := bytes.TrimSpace(args)
	if channelID[0] != '#' {
		return fmt.Errorf("->> ERR: ChannelID must begin with '#'")
	}

	c.outbound <- command{
		recipient: string(channelID),
		sender:    c,
		id:        JOIN,
	}

	return nil
}

func (c *client) leave(args []byte) error {
	channelID := bytes.TrimSpace(args)
	if channelID[0] != '#' {
		return fmt.Errorf("->> ERR: ChannelID must start with '#'")
	}

	c.outbound <- command{
		recipient: string(channelID),
		sender:    c,
		id:        LEAVE,
	}

	return nil
}

func (c *client) msg(args []byte) error {
	args = bytes.TrimSpace(args)
	if args[0] != '#' && args[0] != '@' {
		return fmt.Errorf("->> ERR: Recipient must be a channel ('#name') or a user (''@user)")
	}

	recipient := bytes.Split(args, []byte(" "))[0]
	if len(recipient) == 1 {
		return fmt.Errorf("->> ERR: Recipient must have a name")
	}

	args = bytes.TrimSpace(bytes.TrimPrefix(args, recipient))
	l := bytes.Split(args, MSG_DELIMITER)[0]
	length, err := strconv.Atoi(string(l))
	if err != nil {
		return fmt.Errorf("->> ERR: Body length must be present")
	}
	if length == 0 {
		return fmt.Errorf("->> ERR: Body length must be at least 1")
	}

	padding := len(l) + len(MSG_DELIMITER) // Size of the body length + delimiter
	body := args[padding : padding+length]

	c.outbound <- command{
		recipient: string(recipient),
		sender:    c,
		body:      body,
		id:        MSG,
	}

	return nil
}

func (c *client) files(args []byte) error {
	channelID := bytes.TrimSpace(args)

	if len(channelID) <= 1 {
		return fmt.Errorf("->> ERR : Channel must have a name ('#name')")
	}

	if channelID[0] != '#' {
		return fmt.Errorf("->> ERR: Please provide a channel to look for files ('#name')")
	}

	c.outbound <- command{
		sender:    c,
		recipient: string(channelID),
		id:        FILES,
	}

	return nil
}

func (c *client) send(args []byte) error {
	args = bytes.TrimSpace(args)
	if args[0] != '#' {
		return fmt.Errorf("->> ERR: Recipient must be a channel ('#name')")
	}

	recipient := bytes.Split(args, []byte(" "))[0]
	if len(recipient) == 1 {
		return fmt.Errorf("->> ERR: Recipient must have a name ('#name')")
	}

	args = bytes.TrimSpace(bytes.TrimPrefix(args, recipient))
	filename := bytes.Split(args, []byte(" "))[0]
	if len(filename) == 1 {
		return fmt.Errorf("->> ERR: File must be saved with a name")
	}

	file := bytes.TrimPrefix(args, filename)
	file = bytes.Split(file, FILE_DELIMITER)[1]
	splittedFile := bytes.Split(file, BREAK_LINE_DELIMITER)
	body := bytes.Join(splittedFile, []byte("\n"))

	c.outbound <- command{
		recipient: string(recipient),
		sender:    c,
		header:    filename,
		body:      body,
		id:        SEND,
	}

	return nil
}

func (c *client) get(args []byte) error {
	args = bytes.TrimSpace(args)
	if args[0] != '#' {
		return fmt.Errorf("->> ERR: Recipient must be a channel ('#name')")
	}

	recipient := bytes.Split(args, []byte(" "))[0]
	if len(recipient) == 1 {
		return fmt.Errorf("->> ERR: Recipient must have a name ('#name')")
	}

	filename := bytes.TrimSpace(bytes.TrimPrefix(args, recipient))
	fn := string(filename)

	if len(fn) == 0 {
		return fmt.Errorf("->> ERR: Filename must be provided")
	}

	c.outbound <- command{
		sender:    c,
		recipient: string(recipient),
		body:      filename,
		id:        GET,
	}

	return nil
}

func (c *client) chns() {
	c.outbound <- command{
		sender: c,
		id:     CHNS,
	}
}

func (c *client) usrs() {
	c.outbound <- command{
		sender: c,
		id:     USRS,
	}
}

func (c *client) err(e error) {
	c.conn.Write([]byte("ERR " + e.Error() + "\n"))
}
