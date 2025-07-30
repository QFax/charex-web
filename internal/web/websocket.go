package web

import (
	"charex/internal/extractors"
	"charex/internal/saver"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections
	},
}

type Client struct {
	server *Server
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// Instead of broadcasting, we now handle the message.
		go c.server.handleMessage(c, message)
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mutex.Unlock()
		case message := <-h.broadcast:
			h.mutex.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mutex.Unlock()
		}
	}
}

// sendJSON is a helper to marshal and send a JSON message to the client.
func (c *Client) sendJSON(v interface{}) error {
	msg, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}
	c.send <- msg
	return nil
}

// sendStatus is a helper to send a status update message to the client.
func (c *Client) sendStatus(status, message string) {
	c.sendJSON(OutgoingMessage{
		Type: "status",
		Payload: StatusPayload{
			Status:  status,
			Message: message,
		},
	})
}

// handleMessage is the central message processor.
func (s *Server) handleMessage(c *Client, rawMessage []byte) {
	var msg WebSocketMessage
	if err := json.Unmarshal(rawMessage, &msg); err != nil {
		log.Printf("Error unmarshalling message: %v", err)
		c.sendStatus("error", "Invalid message format.")
		return
	}

	switch msg.Type {
	case "extract_sakura":
		s.handleExtraction(c, msg.Payload, "SakuraFM", s.sakuraExtractor)
	case "extract_janitor":
		s.handleExtraction(c, msg.Payload, "JanitorAI", s.janitorExtractor)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
		c.sendStatus("error", fmt.Sprintf("Unknown message type: %s", msg.Type))
	}
}

func (s *Server) handleExtraction(c *Client, payload json.RawMessage, sourceName string, extractor extractors.Extractor) {
	var urlPayload ExtractURLPayload
	if err := json.Unmarshal(payload, &urlPayload); err != nil {
		c.sendStatus("error", "Invalid payload for extraction.")
		return
	}
	log.Printf("Handling extraction for source: %s, URL: %s", sourceName, urlPayload.URL)

	c.sendStatus("started", fmt.Sprintf("Starting extraction from %s...", urlPayload.URL))

	card, rawData, cardImage, err := extractor.Extract([]byte(urlPayload.URL))
	if err != nil {
		log.Printf("Error during extraction: %v", err)
		c.sendStatus("error", fmt.Sprintf("Extraction failed: %v", err))
		return
	}

	if err := saver.SaveCard(card, rawData, cardImage, sourceName); err != nil {
		log.Printf("Error saving card: %v", err)
		c.sendStatus("error", fmt.Sprintf("Failed to save card: %v", err))
		return
	}

	c.sendStatus("completed", fmt.Sprintf("Successfully extracted and saved %s.", card.Data.Name))

	// Broadcast the new card to all clients
	broadcastMessage, err := json.Marshal(OutgoingMessage{
		Type: "new_card",
		Payload: NewCardPayload{
			Source: sourceName,
			Card:   *card,
		},
	})
	if err != nil {
		log.Printf("Error marshalling broadcast message: %v", err)
		return
	}
	c.hub.broadcast <- broadcastMessage
}

func serveWs(s *Server, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{server: s, hub: s.hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}