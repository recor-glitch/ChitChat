package server

import (
	"ChitChat/internal/shared/application/routes"
	"ChitChat/internal/shared/application/service/db"
	"log"

	"github.com/gin-gonic/gin"
)

func Run() {
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	r := gin.Default()
	routes.SetupRoutes(r)
	log.Println("Server is running on http://localhost:4000")
	err := r.Run(":4000")
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	log.Println("Server stopped")
}
