package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

func InitDB() error {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}
	var err error
	db, err = pgxpool.New(context.Background(), connStr)
	return err
}

func GetDB() *pgxpool.Pool {
	return db
}
