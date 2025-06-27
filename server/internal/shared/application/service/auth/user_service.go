package auth

import (
	"ChitChat/internal/shared/application/service/db"
	"context"
	"errors"
)

func CreateUser(username, passwordHash string) (*db.User, error) {
	row := db.GetDB().QueryRow(context.Background(),
		`INSERT INTO users (id, email, name, role, password_hash, tenant_id) VALUES (gen_random_uuid(), $1, $1, 'user', $2, (SELECT id FROM tenants LIMIT 1)) RETURNING email, password_hash`,
		username, passwordHash,
	)
	var user db.User
	if err := row.Scan(&user.Username, &user.PasswordHash); err != nil {
		return nil, err
	}
	return &user, nil
}

func FindUserByEmail(username string) (*db.User, error) {
	row := db.GetDB().QueryRow(context.Background(),
		`SELECT email, password_hash FROM users WHERE email=$1`,
		username,
	)
	var user db.User
	if err := row.Scan(&user.Username, &user.PasswordHash); err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}
