package handlers

import (
	"ChitChat/internal/shared/application/service/auth"
	"ChitChat/internal/shared/application/service/db"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type PhoneAuthHandlers struct {
	phoneAuthService *auth.PhoneAuthService
}

func NewPhoneAuthHandlers(phoneAuthService *auth.PhoneAuthService) *PhoneAuthHandlers {
	return &PhoneAuthHandlers{
		phoneAuthService: phoneAuthService,
	}
}

// SendVerificationCode sends a verification code to the provided phone number
func (h *PhoneAuthHandlers) SendVerificationCode(c *gin.Context) {
	var req db.PhoneAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validate phone number format (basic validation)
	if len(req.PhoneNumber) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phone number format"})
		return
	}

	err := h.phoneAuthService.SendVerificationCode(req.PhoneNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Verification code sent successfully",
		"phone_number": req.PhoneNumber,
	})
}

// VerifyCode verifies the provided code for a phone number
func (h *PhoneAuthHandlers) VerifyCode(c *gin.Context) {
	var req db.PhoneVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	valid, err := h.phoneAuthService.VerifyCode(req.PhoneNumber, req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Code verified successfully",
		"phone_number": req.PhoneNumber,
	})
}

// PhoneSignup handles phone-based user registration
func (h *PhoneAuthHandlers) PhoneSignup(c *gin.Context) {
	var req db.PhoneSignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// First verify the phone number
	valid, err := h.phoneAuthService.VerifyCode(req.PhoneNumber, req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please verify your phone number first"})
		return
	}

	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification code"})
		return
	}

	// Create the user
	user, err := h.phoneAuthService.CreateUserWithPhone(req.PhoneNumber, req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate JWT token
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("supersecretkey")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":      user.ID,
		"phone_number": req.PhoneNumber,
		"exp":          time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User created successfully",
		"token":   tokenString,
		"user":    user,
	})
}

// PhoneSignin handles phone-based user authentication
func (h *PhoneAuthHandlers) PhoneSignin(c *gin.Context) {
	var req db.PhoneSigninRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Verify the phone number first
	valid, err := h.phoneAuthService.VerifyCode(req.PhoneNumber, req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please verify your phone number first"})
		return
	}

	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification code"})
		return
	}

	// Authenticate the user
	token, err := h.phoneAuthService.AuthenticateWithPhone(req.PhoneNumber)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Authentication successful",
		"token":   token,
	})
}

// UpdatePhoneNumber updates a user's phone number
func (h *PhoneAuthHandlers) UpdatePhoneNumber(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req db.PhoneAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Update the phone number
	err := h.phoneAuthService.UpdateUserPhone(userID.(string), req.PhoneNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update phone number"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Phone number updated successfully",
		"phone_number": req.PhoneNumber,
	})
}

// ResendVerificationCode resends a verification code to the phone number
func (h *PhoneAuthHandlers) ResendVerificationCode(c *gin.Context) {
	var req db.PhoneAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := h.phoneAuthService.SendVerificationCode(req.PhoneNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resend verification code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Verification code resent successfully",
		"phone_number": req.PhoneNumber,
	})
}
