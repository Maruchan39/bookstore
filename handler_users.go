package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Maruchan39/bookstore/internal/auth"
)

func (cfg *Config) handleCreateUser(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Username  string `json:"username"`
		Password  string `json:"password"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}

	if err := decoder.Decode(&params); err != nil {
		log.Printf("error decoding parameters: %s", err)
		respondWithError(w, http.StatusBadRequest, "could not decode parameters")
		return
	}

	if len(strings.TrimSpace(params.FirstName)) < 2 {
		respondWithError(w, http.StatusBadRequest, "firstName must have 2 or more non-whitespace characters")
		return
	}

	if len(strings.TrimSpace(params.LastName)) < 2 {
		respondWithError(w, http.StatusBadRequest, "lastName must have 2 or more non-whitespace characters")
		return
	}

	if !isValidEmail(params.Username) {
		respondWithError(w, http.StatusBadRequest, "username must be a valid email address")
		return
	}

	if !isStrongPassword(params.Password) {
		respondWithError(w, http.StatusBadRequest, "password must be at least 8 characters long and contain a lowercase letter, an uppercase letter, a number and a special character")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("error hashing password: %s", err)
		respondWithError(w, http.StatusInternalServerError, "could not hash password")
		return
	}

	user := User{
		ID:             uuid.New(),
		Username:       params.Username,
		FirstName:      params.FirstName,
		LastName:       params.LastName,
		HashedPassword: hashedPassword,
	}
	result := cfg.db.Create(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			respondWithError(w, http.StatusConflict, "username is already taken")
			return
		}
		log.Printf("error creating user: %s", result.Error)
		respondWithError(w, http.StatusInternalServerError, "could not create user")
		return
	}

	expiresIn := time.Hour

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, expiresIn)
	if err != nil {
		log.Printf("Error generating JWT token: %s", err)
		w.WriteHeader(500)
		return
	}

	type response struct {
		Token string `json:"token"`
	}

	respondWithJSON(w, http.StatusCreated, response{
		Token: token,
	})
}

func (cfg *Config) handleLogin(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}

	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusBadRequest, "could not decode parameters")
		return
	}

	var user User
	result := cfg.db.Where("username = ?", params.Username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
			return
		}
		log.Printf("Error finding user: %s", result.Error)
		respondWithError(w, http.StatusInternalServerError, "could not find user")
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		log.Printf("Error checking password: %s", err)
		respondWithError(w, http.StatusInternalServerError, "could not check password")
		return
	}

	if !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	expiresIn := time.Hour

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, expiresIn)
	if err != nil {
		log.Printf("Error generating JWT token: %s", err)
		respondWithError(w, http.StatusInternalServerError, "could not generate token")
		return
	}

	type response struct {
		Token string `json:"token"`
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: token,
	})
}

func isValidEmail(username string) bool {
	address, err := mail.ParseAddress(username)
	if err != nil {
		return false
	}
	return address.Address == username
}

func isStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var hasLower, hasUpper, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasLower && hasUpper && hasNumber && hasSpecial
}
