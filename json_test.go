package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondWithJSON(t *testing.T) {
	recorder := httptest.NewRecorder()
	respondWithJSON(recorder, http.StatusCreated, map[string]string{"message": "saved"})
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content type = %q, want application/json", got)
	}
	if got := recorder.Body.String(); got != `{"message":"saved"}` {
		t.Fatalf("body = %q", got)
	}
}

func TestRespondWithError(t *testing.T) {
	recorder := httptest.NewRecorder()
	respondWithError(recorder, http.StatusBadRequest, "title is missing")
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content type = %q, want application/json", got)
	}
	var body testErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error != "title is missing" {
		t.Fatalf("error = %q, want title is missing", body.Error)
	}
}
