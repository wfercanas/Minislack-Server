package main

type file struct {
	name string
	size uint
	body []byte
}

func newFile(name string, size uint, body []byte) *file {
	return &file{
		name: name,
		size: size,
		body: body,
	}
}
