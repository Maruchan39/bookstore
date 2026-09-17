package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"github.com/Maruchan39/bookstore/internal/database"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is not set")
	}

	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	dbQueries := database.New(db)

	cfg := Config{
		port:      port,
		dbQueries: dbQueries,
	}

	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr:    ":" + cfg.port,
	}

	mux.HandleFunc("GET /api/books", cfg.handleGetBooks)

	log.Printf("Serving on: http://localhost:%s\n", cfg.port)
	log.Fatal(server.ListenAndServe())

}
