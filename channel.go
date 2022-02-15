package main

type channel struct {
	name    string
	clients map[*client]bool
	files   map[string]*file
}

func newChannel(name string) *channel {
	return &channel{
		name:    name,
		clients: make(map[*client]bool),
		files:   make(map[string]*file),
	}
}

func (c *channel) broadcast(s string, m []byte) {
	msg := append([]byte(s), "["...)
	msg = append(msg, c.name...)
	msg = append(msg, "]: "...)
	msg = append(msg, m...)
	msg = append(msg, '\n')

	for cl := range c.clients {
		cl.conn.Write(msg)
	}
}
