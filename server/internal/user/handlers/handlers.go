package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "User created"})
}

func GetUserByEmail(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "User found by email"})
}

func GetUserByID(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "User found", "id": id})
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "User updated", "id": id})
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "User deleted", "id": id})
}
