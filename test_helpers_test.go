package main

import (
	"testing"
	"time"

	"github.com/Maruchan39/bookstore/internal/auth"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const testJWTSecret = "a-secret-used-only-by-tests"

func newTestConfig(t *testing.T) *Config {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&User{}, &Book{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	return &Config{db: db, jwtSecret: testJWTSecret}
}

func hashPassword(t *testing.T, password string) string {
	t.Helper()

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	return hash
}

type testBookResponse struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Author          string    `json:"author"`
	PublicationDate time.Time `json:"publication_date"`
	Genres          []string  `json:"genres"`
	IsPrivate       bool      `json:"is_private"`
}

type testErrorResponse struct {
	Error string `json:"error"`
}
