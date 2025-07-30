package web

import (
	"charex/internal/core"
	"encoding/json"
)

// WebSocketMessage is a generic message structure to determine the message type.
type WebSocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// ExtractURLPayload is the payload for messages that involve URL-based extraction.
type ExtractURLPayload struct {
	URL string `json:"url"`
}

// OutgoingMessage is a generic structure for messages sent from the server to the client.
type OutgoingMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// StatusPayload is used for sending status updates to the originating client.
type StatusPayload struct {
	Status  string `json:"status"` // e.g., "started", "completed", "error"
	Message string `json:"message"`
}

// NewCardPayload is used for broadcasting a newly created card to all clients.
type NewCardPayload struct {
	Source string             `json:"source"` // e.g., "sakura", "janitor"
	Card   core.TavernCardV2  `json:"card"`
}