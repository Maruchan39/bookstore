package main

import (
	"context"
	"net/http"

	"github.com/Maruchan39/bookstore/internal/auth"
)

type contextKey string

const userIDContextKey contextKey = "userID"

func (cfg *Config) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		bearerToken, err := auth.GetBearerToken(req.Header)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}

		userID, err := auth.ValidateJWT(bearerToken, cfg.jwtSecret)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "invalid bearer token")
			return
		}

		ctx := context.WithValue(req.Context(), userIDContextKey, userID)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}
