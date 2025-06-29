package handlers

import (
	"ChitChat/internal/shared/application/service/auth"
	"ChitChat/internal/shared/application/service/websocket"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type WebSocketHandlers struct {
	wsService *websocket.WebSocketService
}

func NewWebSocketHandlers(wsService *websocket.WebSocketService) *WebSocketHandlers {
	return &WebSocketHandlers{
		wsService: wsService,
	}
}

// HandleWebSocket handles WebSocket connections with authentication
func (h *WebSocketHandlers) HandleWebSocket(c *gin.Context) {
	// Get JWT token from query parameter or Authorization header
	token := c.Query("token")
	if token == "" {
		// Try to get from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "JWT token required"})
		return
	}

	// Validate JWT token
	claims, err := auth.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid JWT token"})
		return
	}

	// Get user ID from claims
	userID, ok := claims["user_id"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
		return
	}

	// Add user ID to query parameters for WebSocket service
	query := c.Request.URL.Query()
	query.Set("user_id", userID)
	c.Request.URL.RawQuery = query.Encode()

	// Handle WebSocket connection
	h.wsService.HandleWebSocket(c.Writer, c.Request)
}

// GetWebSocketStats returns WebSocket connection statistics
func (h *WebSocketHandlers) GetWebSocketStats(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get room ID from query parameter
	roomID := c.Query("room_id")

	stats := gin.H{
		"total_connected_clients": h.wsService.GetConnectedClients(),
	}

	if roomID != "" {
		stats["room_clients"] = h.wsService.GetRoomClients(roomID)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "WebSocket statistics retrieved",
		"stats":   stats,
	})
}
