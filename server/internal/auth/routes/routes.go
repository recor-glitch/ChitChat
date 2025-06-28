package routes

import (
	"ChitChat/internal/auth/handlers"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(r *gin.Engine) {
	authGroup := r.Group("/auth")
	authGroup.POST("/signup", handlers.Signup)
	authGroup.POST("/signin", handlers.Signin)
	authGroup.POST("/logout", handlers.Logout)
	authGroup.POST("/refresh", handlers.RefreshToken)
	authGroup.POST("/forgot-password", handlers.ForgotPassword)
	authGroup.POST("/reset-password", handlers.ResetPassword)
}
