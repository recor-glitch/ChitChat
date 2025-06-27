package server

import (
	"ChitChat/internal/shared/application/routes"
	"log"

	"github.com/gofiber/fiber/v2"
)

func Run() {
	r := fiber.New()
	routes.SetupRoutes(r)
	log.Println("Server is running on http://localhost:4000")
	err := r.Listen(":4000")
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	log.Println("Server stopped")
}
