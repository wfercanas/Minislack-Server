package main

type file struct {
	filename string
	body     []byte
}

func newFile(filename string, body []byte) *file {
	return &file{
		filename: filename,
		body:     body,
	}
}
