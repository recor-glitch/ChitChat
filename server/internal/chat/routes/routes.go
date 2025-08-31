package routes

import (
	"ChitChat/internal/chat/handlers"
	"ChitChat/internal/shared/application/middleware"
	"ChitChat/internal/shared/application/service/chat"
	"ChitChat/internal/shared/application/service/db"
	"ChitChat/internal/shared/application/service/websocket"

	"github.com/gin-gonic/gin"
)

func SetupChatRoutes(r *gin.Engine) {
	// Initialize WebSocket service
	wsService := websocket.NewWebSocketService()

	// Initialize chat service with WebSocket integration
	chatService := chat.NewChatService(db.GetDB(), wsService)
	chatHandlers := handlers.NewChatHandlers(chatService)

	// Initialize WebSocket handlers
	wsHandlers := handlers.NewWebSocketHandlers(wsService)

	chatAuthRoute := r.Group("/chat")
	chatAuthRoute.Use(middleware.JWTAuth())

	// Room management
	chatAuthRoute.GET("/rooms", chatHandlers.GetChatRooms)
	chatAuthRoute.POST("/rooms", chatHandlers.CreateChatRoom)
	chatAuthRoute.GET("/rooms/:id", chatHandlers.GetRoomInfo)

	// Message management
	chatAuthRoute.GET("/rooms/:id/messages", chatHandlers.GetMessagesByRoom)
	chatAuthRoute.POST("/rooms/:id/messages", chatHandlers.SendMessage)

	// Group chat member management
	chatAuthRoute.POST("/rooms/:id/members", chatHandlers.AddMemberToRoom)
	chatAuthRoute.DELETE("/rooms/:id/members", chatHandlers.RemoveMemberFromRoom)

	// Direct message endpoints
	chatAuthRoute.POST("/direct/message", chatHandlers.SendDirectMessage)
	chatAuthRoute.GET("/direct/room/:recipient_id", chatHandlers.GetDirectMessageRoom)
	chatAuthRoute.GET("/direct/messages/:recipient_id", chatHandlers.GetDirectMessages)

	// WebSocket endpoints (no JWT middleware - authentication handled manually)
	r.GET("/chat/ws", wsHandlers.HandleWebSocket)
	chatAuthRoute.GET("/ws/stats", wsHandlers.GetWebSocketStats)
}
