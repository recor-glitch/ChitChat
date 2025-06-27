package auth

import (
	"ChitChat/internal/shared/application/routes/handlers"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(r *gin.Engine) {

	r.POST("/signup", handlers.Signup)
	r.POST("/signin", handlers.Signin)
	r.POST("/logout", handlers.Logout)
	r.POST("/refresh", handlers.RefreshToken)
	r.POST("/forgot-password", handlers.ForgotPassword)
	r.POST("/reset-password", handlers.ResetPassword)
}
