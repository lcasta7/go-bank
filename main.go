package main

import (
	"log"
)

func main() {
	store, err := NewProstgressStore()

	if err != nil {
		log.Fatal("Error initializing the db")
	}

	if err := store.Init(); err != nil {
		log.Fatal("Error creating table ", err)
	}

	server := NewApiServer(":3000", store)
	server.Run()
}
