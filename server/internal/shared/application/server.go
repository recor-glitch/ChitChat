package server

import (
	"ChitChat/internal/auth/handlers"
	"ChitChat/internal/shared/application/routes"
	"ChitChat/internal/shared/application/service/auth"
	"ChitChat/internal/shared/application/service/db"
	"log"

	"github.com/gin-gonic/gin"
)

func Run() {
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize phone auth service
	phoneAuthService := auth.NewPhoneAuthService(db.GetDB())
	phoneAuthHandlers := handlers.NewPhoneAuthHandlers(phoneAuthService)

	r := gin.Default()
	routes.SetupRoutes(r, phoneAuthHandlers)
	log.Println("Server is running on http://localhost:4000")
	err := r.Run(":4000")
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	log.Println("Server stopped")
}
