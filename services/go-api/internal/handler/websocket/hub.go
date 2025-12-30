package websocket

import (
	"sync"

	"devjournal/internal/domain"
)

// Hub maintains the set of active clients and broadcasts messages to rooms
type Hub struct {
	// Registered clients by room
	rooms map[string]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast messages to a room
	broadcast chan *domain.ChatMessage

	// Mutex for thread-safe room access
	mu sync.RWMutex
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *domain.ChatMessage),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient adds a client to a room
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Create room if it doesn't exist
	if _, ok := h.rooms[client.room]; !ok {
		h.rooms[client.room] = make(map[*Client]bool)
	}

	h.rooms[client.room][client] = true

	// Broadcast join message to room
	joinMessage := domain.NewChatMessage(
		client.room,
		client.userID,
		client.userName,
		"has joined the room",
		"join",
	)
	h.broadcastToRoom(client.room, joinMessage)
}

// unregisterClient removes a client from a room
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, ok := h.rooms[client.room]; ok {
		if _, ok := room[client]; ok {
			// Broadcast leave message before removing
			leaveMessage := domain.NewChatMessage(
				client.room,
				client.userID,
				client.userName,
				"has left the room",
				"leave",
			)
			h.broadcastToRoomExcept(client.room, leaveMessage, client)

			delete(room, client)
			close(client.send)

			// Clean up empty rooms
			if len(room) == 0 {
				delete(h.rooms, client.room)
			}
		}
	}
}

// broadcastMessage sends a message to all clients in a room
func (h *Hub) broadcastMessage(message *domain.ChatMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	h.broadcastToRoom(message.Room, message)
}

// broadcastToRoom sends a message to all clients in a specific room (must hold lock)
func (h *Hub) broadcastToRoom(room string, message *domain.ChatMessage) {
	if clients, ok := h.rooms[room]; ok {
		for client := range clients {
			select {
			case client.send <- message:
			default:
				// Client's send buffer is full, close connection
				close(client.send)
				delete(clients, client)
			}
		}
	}
}

// broadcastToRoomExcept sends a message to all clients except one (must hold lock)
func (h *Hub) broadcastToRoomExcept(room string, message *domain.ChatMessage, except *Client) {
	if clients, ok := h.rooms[room]; ok {
		for client := range clients {
			if client == except {
				continue
			}
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(clients, client)
			}
		}
	}
}

// GetRoomClients returns the number of clients in a room
func (h *Hub) GetRoomClients(room string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.rooms[room]; ok {
		return len(clients)
	}
	return 0
}

// GetRooms returns a list of all active rooms
func (h *Hub) GetRooms() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	rooms := make([]string, 0, len(h.rooms))
	for room := range h.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}
