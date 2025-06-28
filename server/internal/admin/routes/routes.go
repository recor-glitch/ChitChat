package routes

import (
	"ChitChat/internal/admin/handlers"
	"ChitChat/internal/shared/application/middleware"

	"github.com/gin-gonic/gin"
)

func SetupAdminRoutes(r *gin.Engine) {
	adminAuthRoute := r.Group("/admin")
	adminAuthRoute.Use(middleware.JWTAuth())
	adminAuthRoute.Use(middleware.AdminAuthMiddleware())
	adminAuthRoute.GET("/users", handlers.GetAllUsers)
	adminAuthRoute.DELETE("/users/:id", handlers.DeleteUserByAdmin)
	adminAuthRoute.GET("/stats", handlers.GetSystemStats)
}
