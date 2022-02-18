package main

import "bytes"

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

func replaceReturns(body []byte) []byte {
	splittedBody := bytes.Split(body, []byte("\n"))
	joinedBody := bytes.Join(splittedBody, []byte("//"))
	joinedBody = append(joinedBody, '\n')
	return joinedBody
}
