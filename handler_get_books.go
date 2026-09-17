package main

import (
	"net/http"
	"time"
)

type bookResponse struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	Price       string    `json:"price"`
	IsAvailable bool      `json:"is_available"`
}

func (cfg *Config) handleGetBooks(w http.ResponseWriter, req *http.Request) {
	books, err := cfg.dbQueries.GetBooks(req.Context())
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	response := make([]bookResponse, 0, len(books))
	for _, book := range books {
		response = append(response, bookResponse{
			ID:          book.ID,
			CreatedAt:   book.CreatedAt,
			UpdatedAt:   book.UpdatedAt,
			Title:       book.Title,
			Author:      book.Author,
			Price:       book.Price,
			IsAvailable: book.IsAvailable,
		})
	}

	respondWithJSON(w, 200, response)
}
