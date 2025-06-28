package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetChatRooms(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Chat rooms retrieved"})
}

func CreateChatRoom(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Chat room created"})
}

func GetMessagesByRoom(c *gin.Context) {
	roomID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Messages retrieved", "room_id": roomID})
}

func SendMessage(c *gin.Context) {
	roomID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Message sent", "room_id": roomID})
}
