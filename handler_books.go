package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type bookResponse struct {
	ID              string    `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Title           string    `json:"title"`
	Author          string    `json:"author"`
	PublicationDate time.Time `json:"publication_date"`
	Genres          []string  `json:"genres"`
	IsPrivate       bool      `json:"is_private"`
}

func (cfg *Config) handleCreateBook(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Title           string   `json:"title"`
		Author          string   `json:"author"`
		PublicationDate string   `json:"publicationDate"`
		Genres          []string `json:"genres"`
		IsPrivate       *bool    `json:"isPrivate"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}

	if err := decoder.Decode(&params); err != nil {
		log.Printf("error decoding parameters: %s", err)
		respondWithError(w, http.StatusBadRequest, "could not decode parameters")
		return
	}

	if len(strings.TrimSpace(params.Title)) < 2 {
		respondWithError(w, http.StatusBadRequest, "title must have 2 or more non-whitespace characters")
		return
	}

	if len(strings.TrimSpace(params.Author)) < 2 {
		respondWithError(w, http.StatusBadRequest, "author must have 2 or more non-whitespace characters")
		return
	}

	publicationDate, err := time.Parse("2006-01-02", params.PublicationDate)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "publicationDate must be a valid date in ISO8601 format (e.g. 2018-01-22)")
		return
	}

	genres := params.Genres
	if genres == nil {
		genres = []string{}
	}

	isPrivate := true
	if params.IsPrivate != nil {
		isPrivate = *params.IsPrivate
	}

	book := Book{
		ID:              uuid.New(),
		Title:           params.Title,
		Author:          params.Author,
		PublicationDate: publicationDate,
		Genres:          genres,
		IsPrivate:       isPrivate,
	}

	if result := cfg.db.Create(&book); result.Error != nil {
		log.Printf("error creating book: %s", result.Error)
		respondWithError(w, http.StatusInternalServerError, "could not create book")
		return
	}

	respondWithJSON(w, http.StatusCreated, bookResponse{
		ID:              book.ID.String(),
		CreatedAt:       book.CreatedAt,
		UpdatedAt:       book.UpdatedAt,
		Title:           book.Title,
		Author:          book.Author,
		PublicationDate: book.PublicationDate,
		Genres:          book.Genres,
		IsPrivate:       book.IsPrivate,
	})
}

func (cfg *Config) handleGetBooks(w http.ResponseWriter, req *http.Request) {
	// TO DO: Get all the books in the app that are either marked as public books or which belong to you.
	var books []Book
	result := cfg.db.Find(&books)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, result.Error.Error())
		return
	}

	response := make([]bookResponse, 0, len(books))
	for _, book := range books {
		response = append(response, bookResponse{
			ID:              book.ID.String(),
			CreatedAt:       book.CreatedAt,
			UpdatedAt:       book.UpdatedAt,
			Title:           book.Title,
			Author:          book.Author,
			PublicationDate: book.PublicationDate,
			Genres:          book.Genres,
			IsPrivate:       book.IsPrivate,
		})
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (cfg *Config) handleGetBook(w http.ResponseWriter, req *http.Request) {
	// TO DO: The book will only be returned if you own it or its owner has marked it as a public book
	bookID, err := uuid.Parse(req.PathValue("bookID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid book ID")
		return
	}

	var book Book
	result := cfg.db.Where("id = ?", bookID).First(&book)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "book not found")
			return
		}
		log.Printf("error getting book: %s", result.Error)
		respondWithError(w, http.StatusInternalServerError, "could not get book")
		return
	}

	respondWithJSON(w, http.StatusOK, bookResponse{
		ID:              book.ID.String(),
		CreatedAt:       book.CreatedAt,
		UpdatedAt:       book.UpdatedAt,
		Title:           book.Title,
		Author:          book.Author,
		PublicationDate: book.PublicationDate,
		Genres:          book.Genres,
		IsPrivate:       book.IsPrivate,
	})
}

func (cfg *Config) handleUpdateBook(w http.ResponseWriter, req *http.Request) {
	// TO DO: to add owner/userID check so that only the owner can update the book
	type parameters struct {
		Title           *string   `json:"title"`
		Author          *string   `json:"author"`
		PublicationDate *string   `json:"publicationDate"`
		Genres          *[]string `json:"genres"`
		IsPrivate       *bool     `json:"isPrivate"`
	}

	bookID, err := uuid.Parse(req.PathValue("bookID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid book ID")
		return
	}

	var book Book
	result := cfg.db.Where("id = ?", bookID).First(&book)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "book not found")
			return
		}
		log.Printf("error getting book: %s", result.Error)
		respondWithError(w, http.StatusInternalServerError, "could not get book")
		return
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		log.Printf("error decoding parameters: %s", err)
		respondWithError(w, http.StatusBadRequest, "could not decode parameters")
		return
	}

	if params.Title != nil {
		if len(strings.TrimSpace(*params.Title)) < 2 {
			respondWithError(w, http.StatusBadRequest, "title must have 2 or more non-whitespace characters")
			return
		}
		book.Title = *params.Title
	}

	if params.Author != nil {
		if len(strings.TrimSpace(*params.Author)) < 2 {
			respondWithError(w, http.StatusBadRequest, "author must have 2 or more non-whitespace characters")
			return
		}
		book.Author = *params.Author
	}

	if params.PublicationDate != nil {
		publicationDate, err := time.Parse("2006-01-02", *params.PublicationDate)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "publicationDate must be a valid date in ISO8601 format (e.g. 2018-01-22)")
			return
		}
		book.PublicationDate = publicationDate
	}

	if params.Genres != nil {
		genres := *params.Genres
		if genres == nil {
			genres = []string{}
		}
		book.Genres = genres
	}

	if params.IsPrivate != nil {
		book.IsPrivate = *params.IsPrivate
	}

	if result := cfg.db.Save(&book); result.Error != nil {
		log.Printf("error updating book: %s", result.Error)
		respondWithError(w, http.StatusInternalServerError, "could not update book")
		return
	}

	respondWithJSON(w, http.StatusOK, bookResponse{
		ID:              book.ID.String(),
		CreatedAt:       book.CreatedAt,
		UpdatedAt:       book.UpdatedAt,
		Title:           book.Title,
		Author:          book.Author,
		PublicationDate: book.PublicationDate,
		Genres:          book.Genres,
		IsPrivate:       book.IsPrivate,
	})
}

func (cfg *Config) handleDeleteBook(w http.ResponseWriter, req *http.Request) {
	// TO DO: to add owner/userID check so that only the owner can delete the book
	bookID, err := uuid.Parse(req.PathValue("bookID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid book ID")
		return
	}

	var book Book
	result := cfg.db.Where("id = ?", bookID).First(&book)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "book not found")
			return
		}
		log.Printf("error getting book: %s", result.Error)
		respondWithError(w, http.StatusInternalServerError, "could not get book")
		return
	}

	if result := cfg.db.Delete(&book); result.Error != nil {
		log.Printf("error deleting book: %s", result.Error)
		respondWithError(w, http.StatusInternalServerError, "could not delete book")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
