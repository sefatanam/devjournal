package websocket

import (
	"encoding/json"
	"log"
	"time"

	"devjournal/internal/domain"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 4096
)

// Client represents a single WebSocket connection
type Client struct {
	hub *Hub

	// The WebSocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan *domain.ChatMessage

	// Room this client belongs to
	room string

	// User information
	userID   string
	userName string
}

// NewClient creates a new Client instance
func NewClient(hub *Hub, conn *websocket.Conn, room, userID, userName string) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan *domain.ChatMessage, 256),
		room:     room,
		userID:   userID,
		userName: userName,
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse incoming message
		var incomingMessage struct {
			Content string `json:"content"`
			Type    string `json:"type"`
		}
		if err := json.Unmarshal(messageBytes, &incomingMessage); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		// Create chat message
		message := domain.NewChatMessage(
			c.room,
			c.userID,
			c.userName,
			incomingMessage.Content,
			"message",
		)

		// Broadcast to room
		c.hub.broadcast <- message
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Write JSON message
			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("Failed to write message: %v", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
