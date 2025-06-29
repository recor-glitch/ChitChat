package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func CORSMiddleware() gin.HandlerFunc {
	return cors.Default()
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			c.Abort()
			return
		}

		jwtSecret := []byte(os.Getenv("JWT_SECRET"))
		if len(jwtSecret) == 0 {
			jwtSecret = []byte("supersecretkey")
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			fmt.Println(claims)
			c.Set("username", claims["username"])
			c.Set("user_id", claims["user_id"])
		}

		c.Next()
	}
}

func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// For now, just check if user exists
		// In a real app, you'd check user role from database
		username, exists := c.Get("username")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Simple admin check - in real app, check user role
		if username == "admin" {
			c.Next()
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
		}
	}
}
