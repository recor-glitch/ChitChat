package routes

import (
	"ChitChat/internal/auth/handlers"
	"ChitChat/internal/shared/application/middleware"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(r *gin.Engine, phoneAuthHandlers *handlers.PhoneAuthHandlers) {
	authGroup := r.Group("/auth")
	authGroup.POST("/signup", handlers.Signup)
	authGroup.POST("/signin", handlers.Signin)
	authGroup.POST("/logout", handlers.Logout)
	authGroup.POST("/refresh", handlers.RefreshToken)
	authGroup.POST("/forgot-password", handlers.ForgotPassword)
	authGroup.POST("/reset-password", handlers.ResetPassword)

	// Phone authentication routes
	authGroup.POST("/phone/send-code", phoneAuthHandlers.SendVerificationCode)
	authGroup.POST("/phone/verify-code", phoneAuthHandlers.VerifyCode)
	authGroup.POST("/phone/signup", phoneAuthHandlers.PhoneSignup)
	authGroup.POST("/phone/signin", phoneAuthHandlers.PhoneSignin)
	authGroup.POST("/phone/resend-code", phoneAuthHandlers.ResendVerificationCode)
	
	// Protected phone routes
	phoneAuthGroup := authGroup.Group("/phone")
	phoneAuthGroup.Use(middleware.JWTAuth())
	phoneAuthGroup.PUT("/update", phoneAuthHandlers.UpdatePhoneNumber)
}
