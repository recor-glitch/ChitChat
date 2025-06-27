package handlers

import (
	"net/http"

	"ChitChat/internal/shared/application/service/auth"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// Test Route
func TestRoute(c *gin.Context) {
	c.String(http.StatusOK, "Welcome to ChitChat!")
}

// Health Check
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

// Public Info
func GetPublicInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Public information"})
}

// User Handlers
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

// Chat Handlers
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

// Admin Handlers
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

// Auth Handlers
func Signup(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	_, err = auth.CreateUser(req.Username, string(hash))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": req.Username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("supersecretkey")
	}
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func Signin(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	user, err := auth.FindUserByEmail(req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": req.Username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("supersecretkey")
	}
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func RefreshToken(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Token refreshed"})
}

func ForgotPassword(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Password reset email sent"})
}

func ResetPassword(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}
