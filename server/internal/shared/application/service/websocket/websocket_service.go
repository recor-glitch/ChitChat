package websocket

import (
	"ChitChat/internal/shared/application/service/db"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message represents a WebSocket message
type Message struct {
	Type    string      `json:"type"`
	RoomID  string      `json:"room_id,omitempty"`
	Content interface{} `json:"content"`
	UserID  string      `json:"user_id,omitempty"`
}

// Client represents a WebSocket client connection
type Client struct {
	ID     string
	UserID string
	Conn   *websocket.Conn
	Send   chan []byte
	Rooms  map[string]bool // rooms the client is subscribed to
	mu     sync.RWMutex
}

// WebSocketService manages WebSocket connections and message broadcasting
type WebSocketService struct {
	clients  map[string]*Client            // clientID -> client
	rooms    map[string]map[string]*Client // roomID -> clients
	mu       sync.RWMutex
	upgrader websocket.Upgrader
}

// NewWebSocketService creates a new WebSocket service
func NewWebSocketService() *WebSocketService {
	return &WebSocketService{
		clients: make(map[string]*Client),
		rooms:   make(map[string]map[string]*Client),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
	}
}

// HandleWebSocket handles WebSocket connections
func (ws *WebSocketService) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get user ID from query parameter or header
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create new client
	client := &Client{
		ID:     fmt.Sprintf("client_%s_%d", userID, time.Now().UnixNano()),
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Rooms:  make(map[string]bool),
	}

	// Register client
	ws.registerClient(client)

	// Start goroutines for reading and writing
	go ws.readPump(client)
	go ws.writePump(client)
}

// registerClient adds a client to the service
func (ws *WebSocketService) registerClient(client *Client) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.clients[client.ID] = client
	log.Printf("Client %s connected (User: %s)", client.ID, client.UserID)
}

// unregisterClient removes a client from the service
func (ws *WebSocketService) unregisterClient(client *Client) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Check if client is already unregistered
	if _, exists := ws.clients[client.ID]; !exists {
		return
	}

	// Remove from all rooms
	for roomID := range client.Rooms {
		ws.leaveRoom(client, roomID)
	}

	// Remove from clients map
	delete(ws.clients, client.ID)

	// Close connection and channel safely
	client.Conn.Close()

	// Only close channel if it's not already closed
	select {
	case <-client.Send:
		// Channel is already closed
	default:
		close(client.Send)
	}

	log.Printf("Client %s disconnected (User: %s)", client.ID, client.UserID)
}

// SubscribeToRoom adds a client to a room
func (ws *WebSocketService) SubscribeToRoom(clientID, roomID string) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	client, exists := ws.clients[clientID]
	if !exists {
		return fmt.Errorf("client not found")
	}

	// Add client to room
	client.mu.Lock()
	client.Rooms[roomID] = true
	client.mu.Unlock()

	// Add room to rooms map if it doesn't exist
	if ws.rooms[roomID] == nil {
		ws.rooms[roomID] = make(map[string]*Client)
	}
	ws.rooms[roomID][clientID] = client

	log.Printf("Client %s subscribed to room %s", clientID, roomID)
	return nil
}

// UnsubscribeFromRoom removes a client from a room
func (ws *WebSocketService) UnsubscribeFromRoom(clientID, roomID string) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	client, exists := ws.clients[clientID]
	if !exists {
		return fmt.Errorf("client not found")
	}

	ws.leaveRoom(client, roomID)
	return nil
}

// leaveRoom removes a client from a specific room
func (ws *WebSocketService) leaveRoom(client *Client, roomID string) {
	// Remove from client's rooms
	client.mu.Lock()
	delete(client.Rooms, roomID)
	client.mu.Unlock()

	// Remove from room's clients
	if roomClients, exists := ws.rooms[roomID]; exists {
		delete(roomClients, client.ID)

		// Remove room if empty
		if len(roomClients) == 0 {
			delete(ws.rooms, roomID)
		}
	}

	log.Printf("Client %s left room %s", client.ID, roomID)
}

// BroadcastToRoom sends a message to all clients in a room
func (ws *WebSocketService) BroadcastToRoom(roomID string, messageType string, content interface{}, excludeUserID string) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	roomClients, exists := ws.rooms[roomID]
	if !exists {
		return
	}

	message := Message{
		Type:    messageType,
		RoomID:  roomID,
		Content: content,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	for clientID, client := range roomClients {
		// Skip if user should be excluded
		if excludeUserID != "" && client.UserID == excludeUserID {
			continue
		}

		select {
		case client.Send <- messageBytes:
		default:
			// Channel is full, remove client
			log.Printf("Client %s channel full, removing", clientID)
			go ws.unregisterClient(client)
		}
	}
}

// BroadcastMessage sends a new message to all clients in a room
func (ws *WebSocketService) BroadcastMessage(roomID string, message *db.Message) {
	log.Printf("Broadcasting message to room %s: %s", roomID, message.Content)

	ws.mu.RLock()
	roomClients, exists := ws.rooms[roomID]
	ws.mu.RUnlock()

	if !exists {
		log.Printf("No clients subscribed to room %s", roomID)
		return
	}

	log.Printf("Broadcasting to %d clients in room %s", len(roomClients), roomID)
	ws.BroadcastToRoom(roomID, "new_message", message, "")
}

// readPump handles reading messages from the WebSocket connection
func (ws *WebSocketService) readPump(client *Client) {
	defer func() {
		ws.unregisterClient(client)
	}()

	client.Conn.SetReadLimit(512) // 512 bytes max message size
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageBytes, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Parse message
		var message Message
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Handle message based on type
		ws.handleMessage(client, &message)
	}
}

// writePump handles writing messages to the WebSocket connection
func (ws *WebSocketService) writePump(client *Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		ws.unregisterClient(client)
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (ws *WebSocketService) handleMessage(client *Client, message *Message) {
	switch message.Type {
	case "subscribe":
		// Subscribe to a room
		if roomID, ok := message.Content.(string); ok {
			err := ws.SubscribeToRoom(client.ID, roomID)
			if err != nil {
				ws.sendError(client, "Failed to subscribe to room")
			} else {
				ws.sendMessage(client, "subscribed", map[string]string{
					"room_id": roomID,
					"status":  "success",
				})
			}
		}

	case "unsubscribe":
		// Unsubscribe from a room
		if roomID, ok := message.Content.(string); ok {
			err := ws.UnsubscribeFromRoom(client.ID, roomID)
			if err != nil {
				ws.sendError(client, "Failed to unsubscribe from room")
			} else {
				ws.sendMessage(client, "unsubscribed", map[string]string{
					"room_id": roomID,
					"status":  "success",
				})
			}
		}

	case "ping":
		// Respond to ping
		ws.sendMessage(client, "pong", map[string]string{
			"timestamp": time.Now().Format(time.RFC3339),
		})

	default:
		log.Printf("Unknown message type: %s", message.Type)
	}
}

// sendMessage sends a message to a specific client
func (ws *WebSocketService) sendMessage(client *Client, messageType string, content interface{}) {
	message := Message{
		Type:    messageType,
		Content: content,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	select {
	case client.Send <- messageBytes:
	default:
		// Channel is full, remove client
		go ws.unregisterClient(client)
	}
}

// sendError sends an error message to a client
func (ws *WebSocketService) sendError(client *Client, errorMessage string) {
	ws.sendMessage(client, "error", map[string]string{
		"message": errorMessage,
	})
}

// GetConnectedClients returns the number of connected clients
func (ws *WebSocketService) GetConnectedClients() int {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return len(ws.clients)
}

// GetRoomClients returns the number of clients in a specific room
func (ws *WebSocketService) GetRoomClients(roomID string) int {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	if roomClients, exists := ws.rooms[roomID]; exists {
		return len(roomClients)
	}
	return 0
}
