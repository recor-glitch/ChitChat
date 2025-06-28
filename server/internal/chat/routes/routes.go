package routes

import (
	"ChitChat/internal/chat/handlers"
	"ChitChat/internal/shared/application/middleware"

	"github.com/gin-gonic/gin"
)

func SetupChatRoutes(r *gin.Engine) {
	chatAuthRoute := r.Group("/chat")
	chatAuthRoute.Use(middleware.JWTAuth())
	chatAuthRoute.GET("/rooms", handlers.GetChatRooms)
	chatAuthRoute.POST("/rooms", handlers.CreateChatRoom)
	chatAuthRoute.GET("/rooms/:id/messages", handlers.GetMessagesByRoom)
	chatAuthRoute.POST("/rooms/:id/messages", handlers.SendMessage)
}
