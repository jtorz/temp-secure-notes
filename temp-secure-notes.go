package main

import (
	"log"

	"github.com/jtorz/temp-secure-notes/app/server"
)

func main() {
	s, err := server.NewServer()
	if err != nil {
		log.Fatal(err)
	}
	s.Start()
}
