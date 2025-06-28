package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetAllUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "All users retrieved"})
}

func DeleteUserByAdmin(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "User deleted by admin", "id": id})
}

func GetSystemStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "System stats retrieved"})
}
