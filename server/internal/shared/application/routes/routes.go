package routes

import (
	adminroutes "ChitChat/internal/admin/routes"
	authroutes "ChitChat/internal/auth/routes"
	chatroutes "ChitChat/internal/chat/routes"
	"ChitChat/internal/shared/application/middleware"
	"ChitChat/internal/shared/handlers"
	userroutes "ChitChat/internal/user/routes"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// CORS
	r.Use(middleware.CORSMiddleware())

	// TEST ROUTES
	r.GET("/", handlers.TestRoute)

	// AUTH ROUTES (Public)
	authroutes.SetupAuthRoutes(r)

	// USER ROUTES (Public)
	userroutes.SetupUserRoutes(r)

	// CHAT ROUTES (Authenticated)
	chatroutes.SetupChatRoutes(r)

	// ADMIN ROUTES (Admin Only)
	adminroutes.SetupAdminRoutes(r)

	// PUBLIC ROUTES
	r.GET("/health", handlers.HealthCheck)
	r.GET("/public/info", handlers.GetPublicInfo)
}
