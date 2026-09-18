package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Book struct {
	ID          string `gorm:"type:varchar(36);primaryKey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Title       string  `gorm:"type:varchar(255);not null"`
	Author      string  `gorm:"type:varchar(255);not null"`
	Price       float64 `gorm:"type:decimal(10,2);not null;default:0"`
	IsAvailable bool    `gorm:"not null;default:true"`
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

	log.Printf("Serving on: http://localhost:%s\n", cfg.port)
	log.Fatal(server.ListenAndServe())

}
