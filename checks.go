package main

func (h *hub) userRegistered(username string) bool {
	_, exists := h.clients[username]
	return exists
}
