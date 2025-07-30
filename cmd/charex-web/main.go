package main

import (
	"charex/internal/extractors"
	"charex/internal/web"
	"log"
	"net/http"
	"os"
)

func main() {
	hub := web.NewHub()
	go hub.Run()

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "output"
	}

	// Instantiate the extractors.
	sakuraExtractor := extractors.NewSakuraFMExtractor()
	janitorExtractor := extractors.NewJanitorAIExtractor()

	server := web.NewServer(hub, dataDir, sakuraExtractor, janitorExtractor)

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