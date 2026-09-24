package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	db, err := gorm.Open(mysql.Open(dbUrl), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	if err := db.AutoMigrate(&User{}, &Book{}); err != nil {
		log.Fatal("failed to migrate database")
	}

	cfg := Config{
		port:      port,
		db:        db,
		jwtSecret: jwtSecret,
	}

	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr:    ":" + cfg.port,
	}

	mux.HandleFunc("POST /api/v1/users/signup", cfg.handleCreateUser)
	mux.HandleFunc("POST /api/v1/users/login", cfg.handleLogin)

	mux.Handle("GET /api/v1/users/mybooks", cfg.requireAuth(http.HandlerFunc(cfg.handleGetUserBooks)))

	mux.Handle("POST /api/v1/books", cfg.requireAuth(http.HandlerFunc(cfg.handleCreateBook)))
	mux.Handle("GET /api/v1/books", cfg.requireAuth(http.HandlerFunc(cfg.handleGetBooks)))
	mux.Handle("GET /api/v1/books/{bookID}", cfg.requireAuth(http.HandlerFunc(cfg.handleGetBook)))
	mux.Handle("PUT /api/v1/books/{bookID}", cfg.requireAuth(http.HandlerFunc(cfg.handleUpdateBook)))
	mux.Handle("DELETE /api/v1/books/{bookID}", cfg.requireAuth(http.HandlerFunc(cfg.handleDeleteBook)))

	log.Printf("Serving on: http://localhost:%s\n", cfg.port)
	log.Fatal(server.ListenAndServe())
}
