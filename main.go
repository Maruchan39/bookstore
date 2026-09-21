package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Book struct {
	// to add owner/userID
	ID              uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Title           string
	Author          string
	PublicationDate time.Time
	Genres          []string `gorm:"serializer:json"`
	IsPrivate       bool
}

func main() {
	godotenv.Load(".env")

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("DB_URL environment variable is not set")
	}

	db, err := gorm.Open(mysql.Open(dbUrl), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	if err := db.AutoMigrate(&Book{}); err != nil {
		log.Fatal("failed to migrate database")
	}

	cfg := Config{
		port: port,
		db:   db,
	}

	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr:    ":" + cfg.port,
	}

	mux.HandleFunc("POST /api/books", cfg.handleCreateBook)
	mux.HandleFunc("GET /api/books", cfg.handleGetBooks)
	mux.HandleFunc("GET /api/books/{bookID}", cfg.handleGetBook)
	mux.HandleFunc("PUT /api/books/{bookID}", cfg.handleUpdateBook)
	mux.HandleFunc("DELETE /api/books/{bookID}", cfg.handleDeleteBook)

	log.Printf("Serving on: http://localhost:%s\n", cfg.port)
	log.Fatal(server.ListenAndServe())
}
