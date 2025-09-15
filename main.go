package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	var store Store
	if os.Getenv("FAV_STORE") == "file" {
		path := os.Getenv("FAV_PATH")
		if path == "" {
			path = "data/favourites.json"
		}
		fs, err := NewFileStore(path)
		if err != nil {
			log.Fatalf("failed to init file store: %v", err)
		}
		store = fs
		log.Printf("Using file store at %s", path)
	} else {
		store = NewInMemoryStore()
		log.Printf("Using in-memory store")
	}
	server := NewServer(store)

	addr := ":8080"
	log.Printf("Starting favourites service on %s", addr)
	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
