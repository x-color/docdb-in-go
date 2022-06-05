package main

import (
	"log"

	"github.com/x-color/docdb/server"
)

func main() {
	s := server.NewServer("0.0.0.0", 8080)
	log.Println("Start Server")
	if err := s.Start(); err != nil {
		log.Println(err)
	}
	log.Println("Stop Server")
}
