package routes

import (
	"ChitChat/internal/shared/application/middleware"
	"ChitChat/internal/user/handlers"

	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(r *gin.Engine) {
	r.POST("/user", handlers.CreateUser)
	r.POST("/get-by-email", handlers.GetUserByEmail)

	userAuthRoute := r.Group("/user")
	userAuthRoute.Use(middleware.JWTAuth())
	userAuthRoute.GET(":id", handlers.GetUserByID)
	userAuthRoute.PATCH(":id", handlers.UpdateUser)
	userAuthRoute.DELETE(":id", handlers.DeleteUser)
}
