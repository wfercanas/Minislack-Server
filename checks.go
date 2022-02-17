package main

func (h *hub) userRegistered(username string) bool {
	_, exists := h.clients[username]
	return exists
}

func (h *hub) channelExists(ch string) bool {
	_, exists := h.channels[ch]
	return exists
}
