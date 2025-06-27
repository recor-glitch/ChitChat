package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("supersecretkey")

func RegisterUser(username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = CreateUser(username, string(hash))
	if err != nil {
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)" {
			return errors.New("user already exists")
		}
		return err
	}
	return nil
}

func ValidateUser(username, password string) error {
	user, err := FindUserByEmail(username)
	if err != nil {
		return errors.New("user not found")
	}
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
}

func Authenticate(username, password string) (string, error) {
	if err := ValidateUser(username, password); err != nil {
		return "", errors.New("invalid credentials")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func Signup(username, password string) (string, error) {
	if err := RegisterUser(username, password); err != nil {
		return "", err
	}
	return Authenticate(username, password)
}
