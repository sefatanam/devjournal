package websocket

import (
	"log"
	"net/http"

	"devjournal/internal/middleware"
	"devjournal/internal/service"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, validate the origin
		// For development, allow all origins
		return true
	},
}

// ChatHandler handles WebSocket connections for chat
type ChatHandler struct {
	hub         *Hub
	authService *service.AuthService
}

// NewChatHandler creates a new chat handler
func NewChatHandler(hub *Hub, authService *service.AuthService) *ChatHandler {
	return &ChatHandler{
		hub:         hub,
		authService: authService,
	}
}

// HandleWebSocket handles WebSocket upgrade and connection
func (h *ChatHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get room from path
	room := r.PathValue("room")
	if room == "" {
		http.Error(w, "room is required", http.StatusBadRequest)
		return
	}

	// Get user info from context (set by auth middleware)
	userID := middleware.GetUserID(r.Context())
	userName := middleware.GetUserName(r.Context())
	userEmail := middleware.GetUserEmail(r.Context())

	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Fallback to email prefix if display name is empty
	if userName == "" {
		if userEmail != "" {
			// Use email prefix as display name
			for i, c := range userEmail {
				if c == '@' {
					userName = userEmail[:i]
					break
				}
			}
		}
		if userName == "" {
			userName = "User"
		}
	}

	log.Printf("WebSocket connection: userID=%s, userName=%s, room=%s", userID, userName, room)

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create client
	client := NewClient(h.hub, conn, room, userID, userName)

	// Register client with hub
	h.hub.register <- client

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()
}
