package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Maruchan39/bookstore/internal/auth"
	"github.com/google/uuid"
)

func TestRequireAuth(t *testing.T) {
	cfg := newTestConfig(t)
	userID := uuid.New()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, ok := r.Context().Value(userIDContextKey).(uuid.UUID)
		if !ok || got != userID {
			t.Fatalf("context user ID = %v, %v; want %v, %v", got, ok, userID, true)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	handler := cfg.requireAuth(next)

	for _, tc := range []struct {
		name   string
		header string
		value  string
	}{
		{"no authentication header", "", ""},
		{"authorization header is ignored", "Authorization", "Bearer definitely-not-a-jwt"},
		{"token is not valid", auth.AuthenticationHeader, "definitely-not-a-jwt"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/private", nil)
			if tc.header != "" {
				req.Header.Set(tc.header, tc.value)
			}
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want unauthorized", recorder.Code)
			}
		})
	}

	t.Run("passes the user ID to the next handler", func(t *testing.T) {
		token, err := auth.MakeJWT(userID, cfg.jwtSecret, time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodGet, "/private", nil)
		req.Header.Set(auth.AuthenticationHeader, token)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		if recorder.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want no content", recorder.Code)
		}
	})
}
