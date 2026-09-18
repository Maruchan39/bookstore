package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}

	cfg := Config{
		port: port,
	}

	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr:    ":" + cfg.port,
	}

	log.Printf("Serving on: http://localhost:%s\n", cfg.port)
	log.Fatal(server.ListenAndServe())

}
