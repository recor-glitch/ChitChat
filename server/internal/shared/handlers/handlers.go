package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func TestRoute(c *gin.Context) {
	c.String(http.StatusOK, "Welcome to ChitChat version 1.0!")
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func GetPublicInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Public information"})
}
