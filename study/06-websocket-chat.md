# WebSocket Real-Time Chat

## Overview

DevJournal implements real-time chat for study groups using WebSockets. This enables bidirectional communication between the browser and server without polling.

## How WebSocket Works

```
┌──────────────┐                        ┌──────────────┐
│   Browser    │                        │    Server    │
│   (Client)   │                        │   (Go API)   │
└──────┬───────┘                        └──────┬───────┘
       │                                       │
       │ 1. HTTP Upgrade Request               │
       │ GET /ws/chat/room123                  │
       │ Upgrade: websocket                    │
       │──────────────────────────────────────>│
       │                                       │
       │ 2. HTTP 101 Switching Protocols       │
       │<──────────────────────────────────────│
       │                                       │
       │ ═══════ WebSocket Connection ═══════  │
       │                                       │
       │ 3. Send message                       │
       │ {"type":"message","content":"Hi!"}    │
       │──────────────────────────────────────>│
       │                                       │
       │ 4. Broadcast to all clients           │
       │ {"type":"message","userId":"..."}     │
       │<──────────────────────────────────────│
       │                                       │
```

## Go WebSocket Implementation

### Hub - Connection Manager

```go
// services/go-api/internal/handler/websocket/hub.go

package websocket

import (
    "sync"
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
    // Registered clients by room
    rooms map[string]map[*Client]bool

    // Register requests from clients
    register chan *Client

    // Unregister requests from clients
    unregister chan *Client

    // Inbound messages from clients to broadcast
    broadcast chan *BroadcastMessage

    // Mutex for thread-safe room access
    mu sync.RWMutex
}

// BroadcastMessage wraps a message with its target room
type BroadcastMessage struct {
    RoomID  string
    Message []byte
    Sender  *Client // Optional: exclude sender from broadcast
}

func NewHub() *Hub {
    return &Hub{
        rooms:      make(map[string]map[*Client]bool),
        register:   make(chan *Client),
        unregister: make(chan *Client),
        broadcast:  make(chan *BroadcastMessage),
    }
}

// Run starts the hub's main loop
func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.addClient(client)

        case client := <-h.unregister:
            h.removeClient(client)

        case message := <-h.broadcast:
            h.broadcastToRoom(message)
        }
    }
}

func (h *Hub) addClient(client *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()

    // Create room if it doesn't exist
    if h.rooms[client.roomID] == nil {
        h.rooms[client.roomID] = make(map[*Client]bool)
    }

    h.rooms[client.roomID][client] = true

    // Notify room about new member
    joinMsg := &ChatMessage{
        Type:            "join",
        UserID:          client.userID,
        UserDisplayName: client.displayName,
        RoomID:          client.roomID,
        Timestamp:       time.Now(),
    }

    h.broadcastToRoomUnsafe(client.roomID, joinMsg, nil)
}

func (h *Hub) removeClient(client *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()

    if _, ok := h.rooms[client.roomID]; ok {
        if _, ok := h.rooms[client.roomID][client]; ok {
            delete(h.rooms[client.roomID][client])
            close(client.send)

            // Notify room about member leaving
            leaveMsg := &ChatMessage{
                Type:            "leave",
                UserID:          client.userID,
                UserDisplayName: client.displayName,
                RoomID:          client.roomID,
                Timestamp:       time.Now(),
            }

            h.broadcastToRoomUnsafe(client.roomID, leaveMsg, nil)

            // Clean up empty rooms
            if len(h.rooms[client.roomID]) == 0 {
                delete(h.rooms, client.roomID)
            }
        }
    }
}

func (h *Hub) broadcastToRoom(msg *BroadcastMessage) {
    h.mu.RLock()
    defer h.mu.RUnlock()

    clients, ok := h.rooms[msg.RoomID]
    if !ok {
        return
    }

    for client := range clients {
        // Skip sender if specified
        if msg.Sender != nil && client == msg.Sender {
            continue
        }

        select {
        case client.send <- msg.Message:
        default:
            // Client buffer full, close connection
            close(client.send)
            delete(clients, client)
        }
    }
}

func (h *Hub) broadcastToRoomUnsafe(roomID string, msg *ChatMessage, exclude *Client) {
    data, _ := json.Marshal(msg)

    clients := h.rooms[roomID]
    for client := range clients {
        if client == exclude {
            continue
        }
        select {
        case client.send <- data:
        default:
            close(client.send)
            delete(clients, client)
        }
    }
}

// GetRoomMembers returns the list of users in a room
func (h *Hub) GetRoomMembers(roomID string) []string {
    h.mu.RLock()
    defer h.mu.RUnlock()

    var members []string
    if clients, ok := h.rooms[roomID]; ok {
        for client := range clients {
            members = append(members, client.userID)
        }
    }
    return members
}
```

### Client - Individual Connection

```go
// services/go-api/internal/handler/websocket/client.go

package websocket

import (
    "encoding/json"
    "time"

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
    hub         *Hub
    conn        *websocket.Conn
    send        chan []byte
    roomID      string
    userID      string
    displayName string
}

// ChatMessage represents a message in the chat
type ChatMessage struct {
    ID              string    `json:"id,omitempty"`
    Type            string    `json:"type"` // "message", "join", "leave", "typing"
    Content         string    `json:"content,omitempty"`
    UserID          string    `json:"userId"`
    UserDisplayName string    `json:"userDisplayName"`
    RoomID          string    `json:"roomId"`
    Timestamp       time.Time `json:"timestamp"`
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
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
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err,
                websocket.CloseGoingAway,
                websocket.CloseAbnormalClosure) {
                log.Printf("WebSocket error: %v", err)
            }
            break
        }

        // Parse incoming message
        var incomingMsg struct {
            Type    string `json:"type"`
            Content string `json:"content"`
        }
        if err := json.Unmarshal(message, &incomingMsg); err != nil {
            continue
        }

        // Create broadcast message
        chatMsg := &ChatMessage{
            ID:              uuid.New().String(),
            Type:            incomingMsg.Type,
            Content:         incomingMsg.Content,
            UserID:          c.userID,
            UserDisplayName: c.displayName,
            RoomID:          c.roomID,
            Timestamp:       time.Now(),
        }

        // Broadcast to room
        data, _ := json.Marshal(chatMsg)
        c.hub.broadcast <- &BroadcastMessage{
            RoomID:  c.roomID,
            Message: data,
            Sender:  nil, // Include sender in broadcast
        }
    }
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
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
                // Hub closed the channel
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            w, err := c.conn.NextWriter(websocket.TextMessage)
            if err != nil {
                return
            }
            w.Write(message)

            // Add queued messages to the current WebSocket message
            n := len(c.send)
            for i := 0; i < n; i++ {
                w.Write([]byte{'\n'})
                w.Write(<-c.send)
            }

            if err := w.Close(); err != nil {
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
```

### HTTP Handler

```go
// services/go-api/internal/handler/websocket/chat_handler.go

package websocket

import (
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        // In production, check specific origins
        origin := r.Header.Get("Origin")
        allowedOrigins := []string{
            "http://localhost:4200",
            "http://localhost:4000",
        }
        for _, allowed := range allowedOrigins {
            if origin == allowed {
                return true
            }
        }
        return false
    },
}

type ChatHandler struct {
    hub         *Hub
    authService *service.AuthService
}

func NewChatHandler(hub *Hub, authSvc *service.AuthService) *ChatHandler {
    return &ChatHandler{
        hub:         hub,
        authService: authSvc,
    }
}

// HandleWebSocket handles WebSocket requests from clients
func (h *ChatHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    roomID := chi.URLParam(r, "roomId")
    if roomID == "" {
        http.Error(w, "Room ID required", http.StatusBadRequest)
        return
    }

    // Authenticate via query param (WebSocket can't use headers easily)
    token := r.URL.Query().Get("token")
    if token == "" {
        http.Error(w, "Token required", http.StatusUnauthorized)
        return
    }

    // Validate token
    claims, err := h.authService.ValidateToken(token)
    if err != nil {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    // Upgrade HTTP connection to WebSocket
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WebSocket upgrade error: %v", err)
        return
    }

    // Create client
    client := &Client{
        hub:         h.hub,
        conn:        conn,
        send:        make(chan []byte, 256),
        roomID:      roomID,
        userID:      claims.UserID,
        displayName: claims.DisplayName,
    }

    // Register client with hub
    h.hub.register <- client

    // Start goroutines for reading and writing
    go client.writePump()
    go client.readPump()
}
```

### Server Setup

```go
// services/go-api/cmd/api/main.go

func main() {
    // ... other initialization ...

    // Create WebSocket hub
    hub := websocket.NewHub()
    go hub.Run() // Start hub in background

    // Create chat handler
    chatHandler := websocket.NewChatHandler(hub, authService)

    // Add WebSocket route
    r.Get("/ws/chat/{roomId}", chatHandler.HandleWebSocket)

    // Start server
    log.Println("Starting server on :8080")
    http.ListenAndServe(":8080", r)
}
```

## Angular WebSocket Client

### Chat Service

```typescript
// libs/features/chat/src/lib/services/chat-websocket.service.ts

import { Injectable, inject, signal, computed } from '@angular/core';
import { AuthStore } from '@devjournal/feature-auth';
import { ChatMessage, ConnectionStatus } from '@devjournal/shared-models';

@Injectable({ providedIn: 'root' })
export class ChatWebSocketService {
  private readonly authStore = inject(AuthStore);

  private socket: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;

  // Reactive state with signals
  private readonly _messages = signal<ChatMessage[]>([]);
  private readonly _connectionStatus = signal<ConnectionStatus>('disconnected');
  private readonly _currentRoomId = signal<string | null>(null);

  // Public computed signals
  readonly messages = computed(() => this._messages());
  readonly connectionStatus = computed(() => this._connectionStatus());
  readonly currentRoomId = computed(() => this._currentRoomId());
  readonly isConnected = computed(() => this._connectionStatus() === 'connected');

  // Sorted messages (newest last)
  readonly sortedMessages = computed(() => {
    return [...this._messages()].sort(
      (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
    );
  });

  // Connect to a chat room
  connect(roomId: string): void {
    if (this.socket?.readyState === WebSocket.OPEN) {
      this.disconnect();
    }

    this._currentRoomId.set(roomId);
    this._connectionStatus.set('connecting');
    this._messages.set([]);

    const token = this.authStore.token();
    if (!token) {
      this._connectionStatus.set('error');
      return;
    }

    // Build WebSocket URL
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const url = `${protocol}//${host}/ws/chat/${roomId}?token=${token}`;

    try {
      this.socket = new WebSocket(url);
      this.setupEventListeners();
    } catch (error) {
      console.error('WebSocket connection error:', error);
      this._connectionStatus.set('error');
    }
  }

  // Disconnect from current room
  disconnect(): void {
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
    this._connectionStatus.set('disconnected');
    this._currentRoomId.set(null);
  }

  // Send a message
  sendMessage(content: string): void {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      console.error('WebSocket not connected');
      return;
    }

    const message = {
      type: 'message',
      content: content.trim(),
    };

    this.socket.send(JSON.stringify(message));
  }

  // Send typing indicator
  sendTyping(): void {
    if (this.socket?.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify({ type: 'typing' }));
    }
  }

  private setupEventListeners(): void {
    if (!this.socket) return;

    this.socket.onopen = () => {
      console.log('WebSocket connected');
      this._connectionStatus.set('connected');
      this.reconnectAttempts = 0;
    };

    this.socket.onclose = (event) => {
      console.log('WebSocket closed:', event.code, event.reason);

      if (event.code !== 1000) {
        // Abnormal closure, attempt reconnect
        this._connectionStatus.set('disconnected');
        this.attemptReconnect();
      } else {
        this._connectionStatus.set('disconnected');
      }
    };

    this.socket.onerror = (error) => {
      console.error('WebSocket error:', error);
      this._connectionStatus.set('error');
    };

    this.socket.onmessage = (event) => {
      try {
        const message: ChatMessage = JSON.parse(event.data);
        this.handleMessage(message);
      } catch (error) {
        console.error('Failed to parse message:', error);
      }
    };
  }

  private handleMessage(message: ChatMessage): void {
    // Add message to state
    this._messages.update((messages) => [...messages, message]);

    // Log system messages
    if (message.type === 'join') {
      console.log(`${message.userDisplayName} joined the chat`);
    } else if (message.type === 'leave') {
      console.log(`${message.userDisplayName} left the chat`);
    }
  }

  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.log('Max reconnect attempts reached');
      this._connectionStatus.set('error');
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

    console.log(`Attempting reconnect in ${delay}ms (attempt ${this.reconnectAttempts})`);

    setTimeout(() => {
      const roomId = this._currentRoomId();
      if (roomId) {
        this.connect(roomId);
      }
    }, delay);
  }
}
```

### Chat Store

```typescript
// libs/features/chat/src/lib/store/chat.store.ts

import { computed, inject } from '@angular/core';
import {
  signalStore,
  withState,
  withComputed,
  withMethods,
  patchState,
} from '@ngrx/signals';
import { rxMethod } from '@ngrx/signals/rxjs-interop';
import { pipe, switchMap, tap } from 'rxjs';
import { tapResponse } from '@ngrx/operators';
import { ChatWebSocketService } from '../services/chat-websocket.service';
import { StudyGroupApiService } from '@devjournal/data-access-api';
import {
  StudyGroup,
  GroupMember,
  ChatMessage,
  ConnectionStatus,
} from '@devjournal/shared-models';

interface ChatState {
  myGroups: StudyGroup[];
  publicGroups: StudyGroup[];
  currentRoom: StudyGroup | null;
  members: GroupMember[];
  isLoading: boolean;
  error: string | null;
}

const initialState: ChatState = {
  myGroups: [],
  publicGroups: [],
  currentRoom: null,
  members: [],
  isLoading: false,
  error: null,
};

export const ChatStore = signalStore(
  { providedIn: 'root' },

  withState(initialState),

  withComputed((store, wsService = inject(ChatWebSocketService)) => ({
    // Delegate to WebSocket service signals
    messages: computed(() => wsService.messages()),
    sortedMessages: computed(() => wsService.sortedMessages()),
    connectionStatus: computed(() => wsService.connectionStatus()),
    isConnected: computed(() => wsService.isConnected()),

    // Count my groups
    myGroupsCount: computed(() => store.myGroups().length),

    // Online members (would need additional tracking)
    onlineCount: computed(() => store.members().length),
  })),

  withMethods((
    store,
    wsService = inject(ChatWebSocketService),
    groupApi = inject(StudyGroupApiService)
  ) => ({
    // Load user's groups
    loadMyGroups: rxMethod<void>(
      pipe(
        tap(() => patchState(store, { isLoading: true, error: null })),
        switchMap(() =>
          groupApi.listMyGroups().pipe(
            tapResponse({
              next: (groups) => {
                patchState(store, { myGroups: groups, isLoading: false });
              },
              error: (err: Error) => {
                patchState(store, { error: err.message, isLoading: false });
              },
            })
          )
        )
      )
    ),

    // Load public groups
    loadPublicGroups: rxMethod<void>(
      pipe(
        tap(() => patchState(store, { isLoading: true })),
        switchMap(() =>
          groupApi.listPublicGroups().pipe(
            tapResponse({
              next: (groups) => {
                patchState(store, { publicGroups: groups, isLoading: false });
              },
              error: (err: Error) => {
                patchState(store, { error: err.message, isLoading: false });
              },
            })
          )
        )
      )
    ),

    // Connect to a room
    connectToRoom(roomId: string) {
      const room = store.myGroups().find((g) => g.id === roomId);
      patchState(store, { currentRoom: room || null });
      wsService.connect(roomId);
    },

    // Disconnect from room
    disconnectFromRoom() {
      wsService.disconnect();
      patchState(store, { currentRoom: null, members: [] });
    },

    // Send a message
    sendMessage(content: string) {
      wsService.sendMessage(content);
    },

    // Load room members
    loadMembers: rxMethod<string>(
      pipe(
        switchMap((roomId) =>
          groupApi.getMembers(roomId).pipe(
            tapResponse({
              next: (members) => {
                patchState(store, { members });
              },
              error: () => {},
            })
          )
        )
      )
    ),

    // Create a new group
    createGroup: rxMethod<{ name: string; description: string; isPublic: boolean }>(
      pipe(
        tap(() => patchState(store, { isLoading: true, error: null })),
        switchMap((data) =>
          groupApi.create(data).pipe(
            tapResponse({
              next: (group) => {
                patchState(store, {
                  myGroups: [...store.myGroups(), group],
                  isLoading: false,
                });
              },
              error: (err: Error) => {
                patchState(store, { error: err.message, isLoading: false });
              },
            })
          )
        )
      )
    ),

    // Join a group
    joinGroup: rxMethod<string>(
      pipe(
        tap(() => patchState(store, { isLoading: true })),
        switchMap((groupId) =>
          groupApi.join(groupId).pipe(
            tapResponse({
              next: () => {
                // Move from public to my groups
                const group = store.publicGroups().find((g) => g.id === groupId);
                if (group) {
                  patchState(store, {
                    myGroups: [...store.myGroups(), group],
                    publicGroups: store.publicGroups().filter((g) => g.id !== groupId),
                    isLoading: false,
                  });
                }
              },
              error: (err: Error) => {
                patchState(store, { error: err.message, isLoading: false });
              },
            })
          )
        )
      )
    ),

    // Delete a group
    deleteGroup: rxMethod<string>(
      pipe(
        tap(() => patchState(store, { isLoading: true })),
        switchMap((groupId) =>
          groupApi.delete(groupId).pipe(
            tapResponse({
              next: () => {
                patchState(store, {
                  myGroups: store.myGroups().filter((g) => g.id !== groupId),
                  isLoading: false,
                });
              },
              error: (err: Error) => {
                patchState(store, { error: err.message, isLoading: false });
              },
            })
          )
        )
      )
    ),
  }))
);
```

### Chat Room Component

```typescript
// libs/features/chat/src/lib/components/chat-room/chat-room.component.ts

import { Component, inject, input, effect, ViewChild, ElementRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ChatStore } from '../../store/chat.store';

@Component({
  selector: 'lib-chat-room',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="chat-room">
      <header class="room-header">
        <div class="room-info">
          <h2>{{ store.currentRoom()?.name || 'Chat Room' }}</h2>
          <span class="connection-status" [class]="store.connectionStatus()">
            {{ store.connectionStatus() }}
          </span>
        </div>
        <div class="room-members">
          <span class="member-count">
            {{ store.members().length }} members
          </span>
        </div>
      </header>

      <div class="messages-container" #messagesContainer>
        @if (store.sortedMessages().length === 0) {
          <div class="empty-messages">
            <p>No messages yet. Start the conversation!</p>
          </div>
        } @else {
          @for (message of store.sortedMessages(); track message.id) {
            <div
              class="message"
              [class.system]="message.type !== 'message'"
              [class.own]="isOwnMessage(message.userId)"
            >
              @if (message.type === 'message') {
                <div class="message-header">
                  <span class="message-author">{{ message.userDisplayName }}</span>
                  <span class="message-time">{{ formatTime(message.timestamp) }}</span>
                </div>
                <div class="message-content">{{ message.content }}</div>
              } @else {
                <div class="system-message">
                  @switch (message.type) {
                    @case ('join') {
                      <span>{{ message.userDisplayName }} joined the chat</span>
                    }
                    @case ('leave') {
                      <span>{{ message.userDisplayName }} left the chat</span>
                    }
                  }
                </div>
              }
            </div>
          }
        }
      </div>

      <footer class="message-input-container">
        @if (!store.isConnected()) {
          <div class="connecting-overlay">
            @if (store.connectionStatus() === 'connecting') {
              <span>Connecting...</span>
            } @else if (store.connectionStatus() === 'error') {
              <span>Connection error. Please try again.</span>
              <button (click)="reconnect()">Retry</button>
            }
          </div>
        }
        <input
          type="text"
          [(ngModel)]="messageText"
          (keydown.enter)="sendMessage()"
          placeholder="Type a message..."
          [disabled]="!store.isConnected()"
        />
        <button
          (click)="sendMessage()"
          [disabled]="!store.isConnected() || !messageText.trim()"
        >
          Send
        </button>
      </footer>
    </div>
  `,
})
export class ChatRoomComponent {
  readonly roomId = input.required<string>();
  readonly store = inject(ChatStore);

  @ViewChild('messagesContainer') messagesContainer!: ElementRef;

  messageText = '';

  constructor() {
    // Connect when roomId changes
    effect(() => {
      const id = this.roomId();
      if (id) {
        this.store.connectToRoom(id);
        this.store.loadMembers(id);
      }
    });

    // Auto-scroll when new messages arrive
    effect(() => {
      const messages = this.store.sortedMessages();
      if (messages.length > 0) {
        setTimeout(() => this.scrollToBottom(), 0);
      }
    });
  }

  sendMessage(): void {
    if (this.messageText.trim() && this.store.isConnected()) {
      this.store.sendMessage(this.messageText);
      this.messageText = '';
    }
  }

  reconnect(): void {
    this.store.connectToRoom(this.roomId());
  }

  isOwnMessage(userId: string): boolean {
    // Compare with current user ID from auth store
    return false; // Implement based on auth store
  }

  formatTime(date: Date): string {
    return new Date(date).toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  private scrollToBottom(): void {
    if (this.messagesContainer) {
      const el = this.messagesContainer.nativeElement;
      el.scrollTop = el.scrollHeight;
    }
  }
}
```

## Message Types

```typescript
// libs/shared/models/src/lib/chat.model.ts

export type MessageType = 'message' | 'join' | 'leave' | 'typing';

export interface ChatMessage {
  id: string;
  type: MessageType;
  content?: string;
  userId: string;
  userDisplayName: string;
  roomId: string;
  timestamp: Date;
}

export type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'error';
```

## Key Concepts

### 1. WebSocket vs HTTP Polling

| WebSocket | HTTP Polling |
|-----------|-------------|
| Persistent connection | New request each time |
| Bidirectional | Request-response only |
| Low latency | High latency |
| Server can push | Client must ask |
| More complex | Simple to implement |

### 2. Hub Pattern
Central manager that:
- Tracks all connected clients
- Routes messages to correct rooms
- Handles join/leave notifications
- Manages connection lifecycle

### 3. Goroutines for Each Client
- `readPump` - Reads messages from WebSocket
- `writePump` - Writes messages to WebSocket
- Both run concurrently, communicate via channels

### 4. Heartbeat (Ping/Pong)
Keep connections alive and detect dead connections:
- Server sends ping every 54 seconds
- Client must respond with pong within 60 seconds
- No pong = connection dead

### 5. Reconnection Strategy
Exponential backoff for reconnection:
- Attempt 1: 1 second delay
- Attempt 2: 2 seconds delay
- Attempt 3: 4 seconds delay
- ...up to max attempts

## Best Practices

1. **Authenticate on connect** - Validate JWT in query param
2. **Handle disconnections gracefully** - Auto-reconnect with backoff
3. **Use channels for concurrency** - Don't share state between goroutines
4. **Limit message size** - Prevent DoS attacks
5. **Buffer messages** - Use buffered channels to prevent blocking
6. **Clean up resources** - Close channels and connections properly

## Next Steps

- [Authentication & JWT](./07-authentication.md) - Token validation for WebSocket
- [Angular Signal Store](./05-signal-store.md) - Chat store with signals
