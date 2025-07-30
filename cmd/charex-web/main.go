package main

import (
	"charex/internal/web"
	"log"
	"net/http"
	"os"
)

func main() {
	hub := web.NewHub()
	go hub.Run()

	server := web.NewServer(hub)

	http.HandleFunc("/ws", server.ServeWs)
	http.HandleFunc("/api/cards", server.GetCards)
	http.Handle("/", http.FileServer(http.Dir("./web/static")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "9111"
	}

	log.Printf("Server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("could not listen on port %s %v", port, err)
	}
}