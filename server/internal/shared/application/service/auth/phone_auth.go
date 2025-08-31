package auth

import (
	"ChitChat/internal/shared/application/service/db"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PhoneAuthService struct {
	db         *pgxpool.Pool
	smsService SMSService
}

func NewPhoneAuthService(database *pgxpool.Pool) *PhoneAuthService {
	return &PhoneAuthService{
		db:         database,
		smsService: GetSMSService(),
	}
}

// generateVerificationCode generates a 6-digit verification code
func (s *PhoneAuthService) generateVerificationCode() (string, error) {
	code := ""
	for i := 0; i < 6; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += fmt.Sprintf("%d", num.Int64())
	}
	return code, nil
}

// SendVerificationCode sends a verification code to the phone number
func (s *PhoneAuthService) SendVerificationCode(phoneNumber string) error {
	ctx := context.Background()

	// Generate a 6-digit verification code
	code, err := s.generateVerificationCode()
	if err != nil {
		return err
	}

	// Set expiration time (10 minutes from now)
	expiresAt := time.Now().Add(10 * time.Minute)

	// Store the verification code in the database
	_, err = s.db.Exec(ctx, `
		INSERT INTO phone_verification_codes (phone_number, code, expires_at)
		VALUES ($1, $2, $3)
	`, phoneNumber, code, expiresAt)
	if err != nil {
		return err
	}

	// Send SMS with verification code
	message := fmt.Sprintf("Your ChitChat verification code is: %s", code)
	err = s.smsService.SendSMS(phoneNumber, message)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %v", err)
	}

	return nil
}

// VerifyCode verifies the provided code for a phone number
func (s *PhoneAuthService) VerifyCode(phoneNumber, code string) (bool, error) {
	ctx := context.Background()

	var verificationCode db.PhoneVerificationCode
	err := s.db.QueryRow(ctx, `
		SELECT id, phone_number, code, expires_at, used
		FROM phone_verification_codes
		WHERE phone_number = $1 AND code = $2
		ORDER BY created_at DESC
		LIMIT 1
	`, phoneNumber, code).Scan(&verificationCode.ID, &verificationCode.PhoneNumber, &verificationCode.Code, &verificationCode.ExpiresAt, &verificationCode.Used)

	if err != nil {
		return false, errors.New("invalid verification code")
	}

	// Check if code is expired
	if time.Now().After(verificationCode.ExpiresAt) {
		return false, errors.New("verification code expired")
	}

	// Check if code is already used
	if verificationCode.Used {
		return false, errors.New("verification code already used")
	}

	// Mark the code as used
	_, err = s.db.Exec(ctx, `
		UPDATE phone_verification_codes
		SET used = true
		WHERE id = $1
	`, verificationCode.ID)
	if err != nil {
		return false, err
	}

	return true, nil
}

// CreateUserWithPhone creates a new user with phone number
func (s *PhoneAuthService) CreateUserWithPhone(phoneNumber, name string) (*db.User, error) {
	ctx := context.Background()

	// Check if user already exists with this phone number
	var existingUser db.User
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, phone_number, name, role, phone_verified, created_at
		FROM users WHERE phone_number = $1
	`, phoneNumber).Scan(&existingUser.ID, &existingUser.TenantID, &existingUser.PhoneNumber, &existingUser.Name, &existingUser.Role, &existingUser.PhoneVerified, &existingUser.CreatedAt)

	if err == nil {
		return nil, errors.New("user with this phone number already exists")
	}

	// Get default tenant ID
	var tenantID string
	err = s.db.QueryRow(ctx, "SELECT id FROM tenants LIMIT 1").Scan(&tenantID)
	if err != nil {
		return nil, err
	}

	// Create new user
	var user db.User
	err = s.db.QueryRow(ctx, `
		INSERT INTO users (id, tenant_id, phone_number, name, role, phone_verified)
		VALUES (gen_random_uuid(), $1, $2, $3, 'user', true)
		RETURNING id, tenant_id, phone_number, name, role, phone_verified, created_at
	`, tenantID, phoneNumber, name).Scan(&user.ID, &user.TenantID, &user.PhoneNumber, &user.Name, &user.Role, &user.PhoneVerified, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// FindUserByPhone finds a user by phone number
func (s *PhoneAuthService) FindUserByPhone(phoneNumber string) (*db.User, error) {
	ctx := context.Background()

	var user db.User
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, phone_number, name, role, phone_verified, created_at
		FROM users WHERE phone_number = $1
	`, phoneNumber).Scan(&user.ID, &user.TenantID, &user.PhoneNumber, &user.Name, &user.Role, &user.PhoneVerified, &user.CreatedAt)

	if err != nil {
		return nil, errors.New("user not found")
	}

	return &user, nil
}

// AuthenticateWithPhone authenticates a user with phone number and returns a JWT token
func (s *PhoneAuthService) AuthenticateWithPhone(phoneNumber string) (string, error) {
	user, err := s.FindUserByPhone(phoneNumber)
	if err != nil {
		return "", errors.New("user not found")
	}

	if !user.PhoneVerified {
		return "", errors.New("phone number not verified")
	}

	// Generate JWT token
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("supersecretkey")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":      user.ID,
		"phone_number": phoneNumber,
		"exp":          time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// UpdateUserPhone updates a user's phone number
func (s *PhoneAuthService) UpdateUserPhone(userID, phoneNumber string) error {
	ctx := context.Background()

	_, err := s.db.Exec(ctx, `
		UPDATE users 
		SET phone_number = $1, phone_verified = false
		WHERE id = $2
	`, phoneNumber, userID)

	return err
}

// MarkPhoneAsVerified marks a user's phone number as verified
func (s *PhoneAuthService) MarkPhoneAsVerified(userID string) error {
	ctx := context.Background()

	_, err := s.db.Exec(ctx, `
		UPDATE users 
		SET phone_verified = true
		WHERE id = $1
	`, userID)

	return err
}
