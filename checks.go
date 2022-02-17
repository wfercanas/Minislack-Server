package main

func (h *hub) userRegistered(username string) bool {
	_, exists := h.clients[username]
	return exists
}

func (h *hub) channelExists(ch string) bool {
	_, exists := h.channels[ch]
	return exists
}

func (h *hub) userIsMember(ch *channel, cl *client) bool {
	_, exists := ch.clients[cl]
	return exists
}
