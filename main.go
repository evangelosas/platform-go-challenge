package main

import (
	"log"
	"net/http"
)

func main() {
	store := NewInMemoryStore()
	server := NewServer(store)

	addr := ":8080"
	log.Printf("Starting favourites service on %s", addr)
	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
