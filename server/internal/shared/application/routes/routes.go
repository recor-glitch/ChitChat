package routes

import (
	"ChitChat/internal/shared/application/routes/auth"
	"ChitChat/internal/shared/application/routes/handlers"
	"ChitChat/internal/shared/application/routes/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// CORS
	r.Use(middleware.CORSMiddleware())

	// TEST ROUTES
	r.GET("/", handlers.TestRoute)

	// AUTH ROUTES (Public)
	auth.SetupAuthRoutes(r)

	// USER ROUTES (Public)
	r.POST("/user", handlers.CreateUser)
	r.POST("/get-by-email", handlers.GetUserByEmail)

	// USER ROUTES (Authenticated)
	userAuthRoute := r.Group("/user")
	userAuthRoute.Use(middleware.JWTAuth())
	userAuthRoute.GET("/:id", handlers.GetUserByID)
	userAuthRoute.PATCH("/:id", handlers.UpdateUser)
	userAuthRoute.DELETE("/:id", handlers.DeleteUser)

	// CHAT ROUTES (Authenticated)
	chatAuthRoute := r.Group("/chat")
	chatAuthRoute.Use(middleware.JWTAuth())
	chatAuthRoute.GET("/rooms", handlers.GetChatRooms)
	chatAuthRoute.POST("/rooms", handlers.CreateChatRoom)
	chatAuthRoute.GET("/rooms/:id/messages", handlers.GetMessagesByRoom)
	chatAuthRoute.POST("/rooms/:id/messages", handlers.SendMessage)

	// ADMIN ROUTES (Admin Only)
	adminAuthRoute := r.Group("/admin")
	adminAuthRoute.Use(middleware.JWTAuth())
	adminAuthRoute.Use(middleware.AdminAuthMiddleware())
	adminAuthRoute.GET("/users", handlers.GetAllUsers)
	adminAuthRoute.DELETE("/users/:id", handlers.DeleteUserByAdmin)
	adminAuthRoute.GET("/stats", handlers.GetSystemStats)

	// PUBLIC ROUTES
	r.GET("/health", handlers.HealthCheck)
	r.GET("/public/info", handlers.GetPublicInfo)
}
