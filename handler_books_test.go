package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHandleCreateBook(t *testing.T) {
	cfg := newTestConfig(t)
	owner := User{
		ID:             uuid.New(),
		Username:       "mira@quietcorner.test",
		FirstName:      "Mira",
		LastName:       "Stone",
		HashedPassword: hashPassword(t, "GardenGate7$"),
	}
	if err := cfg.db.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}

	input := map[string]any{
		"title":           "Dievų miškas",
		"author":          "Balys Sruoga",
		"publicationDate": "2021-04-18",
		"genres":          []string{"romanas", "lietuvių literatūra"},
	}
	payload, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(payload))
	req = req.WithContext(context.WithValue(req.Context(), userIDContextKey, owner.ID))
	recorder := httptest.NewRecorder()
	cfg.handleCreateBook(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s; want created", recorder.Code, recorder.Body.String())
	}
	var response testBookResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Title != input["title"] || response.Author != input["author"] || response.IsPrivate != true {
		t.Fatalf("unexpected response: %+v", response)
	}
	if len(response.Genres) != 2 || response.Genres[0] != "romanas" {
		t.Fatalf("genres = %#v", response.Genres)
	}

	var saved Book
	if err := cfg.db.First(&saved, "id = ?", response.ID).Error; err != nil {
		t.Fatal(err)
	}
	wantDate := time.Date(2021, 4, 18, 0, 0, 0, 0, time.UTC)
	if saved.UserID != owner.ID || !saved.PublicationDate.Equal(wantDate) {
		t.Fatalf("saved book = %+v", saved)
	}
}

func TestHandleCreateBookRejectsBadInput(t *testing.T) {
	cfg := newTestConfig(t)
	owner := User{ID: uuid.New(), Username: "reader@lantern.test"}
	if err := cfg.db.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		body map[string]any
	}{
		{"one-letter title", map[string]any{"title": "A", "author": "Someone", "publicationDate": "2020-01-01"}},
		{"one-letter author", map[string]any{"title": "Mažasis princas", "author": "X", "publicationDate": "2020-01-01"}},
		{"date is not ISO", map[string]any{"title": "Mažasis princas", "author": "Antoine de Saint-Exupéry", "publicationDate": "vakar"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.body)
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(data))
			req = req.WithContext(context.WithValue(req.Context(), userIDContextKey, owner.ID))
			recorder := httptest.NewRecorder()
			cfg.handleCreateBook(recorder, req)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want bad request", recorder.Code)
			}
		})
	}
}

func TestHandleListBooksHidesPrivateBooks(t *testing.T) {
	cfg := newTestConfig(t)
	owner := User{ID: uuid.New(), Username: "owner@lantern.test"}
	visitor := User{ID: uuid.New(), Username: "visitor@lantern.test"}
	if err := cfg.db.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	if err := cfg.db.Create(&visitor).Error; err != nil {
		t.Fatal(err)
	}
	public := Book{ID: uuid.New(), UserID: owner.ID, Title: "Dievų miškas", Author: "Balys Sruoga", PublicationDate: time.Now(), IsPrivate: false}
	private := Book{ID: uuid.New(), UserID: owner.ID, Title: "Balta drobulė", Author: "Antanas Škėma", PublicationDate: time.Now(), IsPrivate: true}
	if err := cfg.db.Create(&public).Error; err != nil {
		t.Fatal(err)
	}
	if err := cfg.db.Create(&private).Error; err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want int
	}{
		{"owner sees both books", owner.ID, 2},
		{"visitor sees only public books", visitor.ID, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/books", nil)
			req = req.WithContext(context.WithValue(req.Context(), userIDContextKey, tc.id))
			recorder := httptest.NewRecorder()
			cfg.handleGetBooks(recorder, req)
			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d", recorder.Code)
			}
			var books []testBookResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &books); err != nil {
				t.Fatal(err)
			}
			if len(books) != tc.want {
				t.Fatalf("got %d books, want %d", len(books), tc.want)
			}
		})
	}
}

func TestHandleGetBookAllowsOwnerPrivateBook(t *testing.T) {
	cfg := newTestConfig(t)
	owner := User{ID: uuid.New(), Username: "keeper@lantern.test"}
	if err := cfg.db.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	book := Book{ID: uuid.New(), UserID: owner.ID, Title: "Mažasis princas", Author: "Antoine de Saint-Exupéry", PublicationDate: time.Now(), IsPrivate: true}
	if err := cfg.db.Create(&book).Error; err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/books/"+book.ID.String(), nil)
	req.SetPathValue("bookID", book.ID.String())
	req = req.WithContext(context.WithValue(req.Context(), userIDContextKey, owner.ID))
	recorder := httptest.NewRecorder()
	cfg.handleGetBook(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestHandleGetBookDoesNotExposeAnotherUsersPrivateBook(t *testing.T) {
	cfg := newTestConfig(t)
	owner := User{ID: uuid.New(), Username: "keeper@lantern.test"}
	visitor := User{ID: uuid.New(), Username: "guest@lantern.test"}
	if err := cfg.db.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	if err := cfg.db.Create(&visitor).Error; err != nil {
		t.Fatal(err)
	}
	book := Book{ID: uuid.New(), UserID: owner.ID, Title: "Mažasis princas", Author: "Antoine de Saint-Exupéry", PublicationDate: time.Now(), IsPrivate: true}
	if err := cfg.db.Create(&book).Error; err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/books/"+book.ID.String(), nil)
	req.SetPathValue("bookID", book.ID.String())
	req = req.WithContext(context.WithValue(req.Context(), userIDContextKey, visitor.ID))
	recorder := httptest.NewRecorder()
	cfg.handleGetBook(recorder, req)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want not found", recorder.Code)
	}
}

func TestHandleGetBookRejectsBadID(t *testing.T) {
	cfg := newTestConfig(t)
	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/books/nope", nil)
	req.SetPathValue("bookID", "nope")
	req = req.WithContext(context.WithValue(req.Context(), userIDContextKey, userID))
	recorder := httptest.NewRecorder()
	cfg.handleGetBook(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want bad request", recorder.Code)
	}
}

func TestHandleUpdateBookOnlyOwnerCanChangeIt(t *testing.T) {
	cfg := newTestConfig(t)
	owner := User{ID: uuid.New(), Username: "keeper@lantern.test"}
	stranger := User{ID: uuid.New(), Username: "stranger@lantern.test"}
	if err := cfg.db.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	if err := cfg.db.Create(&stranger).Error; err != nil {
		t.Fatal(err)
	}
	book := Book{ID: uuid.New(), UserID: owner.ID, Title: "Žiedų valdovas: Žiedo brolija", Author: "J. R. R. Tolkien", PublicationDate: time.Now(), IsPrivate: true}
	if err := cfg.db.Create(&book).Error; err != nil {
		t.Fatal(err)
	}

	data, _ := json.Marshal(map[string]any{"title": "Mistborn: The Final Empire", "genres": []string{}, "isPrivate": false})
	blocked := httptest.NewRequest(http.MethodPut, "/books/"+book.ID.String(), bytes.NewReader(data))
	blocked.SetPathValue("bookID", book.ID.String())
	blocked = blocked.WithContext(context.WithValue(blocked.Context(), userIDContextKey, stranger.ID))
	blockedResponse := httptest.NewRecorder()
	cfg.handleUpdateBook(blockedResponse, blocked)
	if blockedResponse.Code != http.StatusNotFound {
		t.Fatalf("stranger status = %d, want not found", blockedResponse.Code)
	}

	data, _ = json.Marshal(map[string]any{"title": "Mistborn: The Final Empire", "genres": []string{}, "isPrivate": false})
	allowed := httptest.NewRequest(http.MethodPut, "/books/"+book.ID.String(), bytes.NewReader(data))
	allowed.SetPathValue("bookID", book.ID.String())
	allowed = allowed.WithContext(context.WithValue(allowed.Context(), userIDContextKey, owner.ID))
	allowedResponse := httptest.NewRecorder()
	cfg.handleUpdateBook(allowedResponse, allowed)
	if allowedResponse.Code != http.StatusOK {
		t.Fatalf("owner status = %d, body = %s", allowedResponse.Code, allowedResponse.Body.String())
	}
	var updated testBookResponse
	if err := json.Unmarshal(allowedResponse.Body.Bytes(), &updated); err != nil {
		t.Fatal(err)
	}
	if updated.Title != "Mistborn: The Final Empire" || updated.IsPrivate || updated.Genres == nil {
		t.Fatalf("unexpected update: %+v", updated)
	}
}

func TestHandleDeleteBookOnlyOwnerCanDeleteIt(t *testing.T) {
	cfg := newTestConfig(t)
	owner := User{ID: uuid.New(), Username: "keeper@lantern.test"}
	stranger := User{ID: uuid.New(), Username: "stranger@lantern.test"}
	if err := cfg.db.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	if err := cfg.db.Create(&stranger).Error; err != nil {
		t.Fatal(err)
	}
	book := Book{ID: uuid.New(), UserID: owner.ID, Title: "Hobitas", Author: "J. R. R. Tolkien", PublicationDate: time.Now(), IsPrivate: true}
	if err := cfg.db.Create(&book).Error; err != nil {
		t.Fatal(err)
	}

	blocked := httptest.NewRequest(http.MethodDelete, "/books/"+book.ID.String(), nil)
	blocked.SetPathValue("bookID", book.ID.String())
	blocked = blocked.WithContext(context.WithValue(blocked.Context(), userIDContextKey, stranger.ID))
	blockedResponse := httptest.NewRecorder()
	cfg.handleDeleteBook(blockedResponse, blocked)
	if blockedResponse.Code != http.StatusNotFound {
		t.Fatalf("stranger status = %d, want not found", blockedResponse.Code)
	}

	allowed := httptest.NewRequest(http.MethodDelete, "/books/"+book.ID.String(), nil)
	allowed.SetPathValue("bookID", book.ID.String())
	allowed = allowed.WithContext(context.WithValue(allowed.Context(), userIDContextKey, owner.ID))
	allowedResponse := httptest.NewRecorder()
	cfg.handleDeleteBook(allowedResponse, allowed)
	if allowedResponse.Code != http.StatusNoContent {
		t.Fatalf("owner status = %d, want no content", allowedResponse.Code)
	}
	var count int64
	cfg.db.Model(&Book{}).Where("id = ?", book.ID).Count(&count)
	if count != 0 {
		t.Fatal("book was not deleted")
	}
}

func TestHandleGetUserBooks(t *testing.T) {
	cfg := newTestConfig(t)
	owner := User{ID: uuid.New(), Username: "keeper@lantern.test"}
	other := User{ID: uuid.New(), Username: "neighbour@lantern.test"}
	if err := cfg.db.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	if err := cfg.db.Create(&other).Error; err != nil {
		t.Fatal(err)
	}
	for _, book := range []Book{
		{ID: uuid.New(), UserID: owner.ID, Title: "Hobitas", Author: "J. R. R. Tolkien", PublicationDate: time.Now(), IsPrivate: true},
		{ID: uuid.New(), UserID: owner.ID, Title: "Mistborn: The Final Empire", Author: "Brandon Sanderson", PublicationDate: time.Now(), IsPrivate: false},
		{ID: uuid.New(), UserID: other.ID, Title: "Žiedų valdovas: Dvi tvirtovės", Author: "J. R. R. Tolkien", PublicationDate: time.Now(), IsPrivate: false},
	} {
		if err := cfg.db.Create(&book).Error; err != nil {
			t.Fatal(err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/mybooks", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDContextKey, owner.ID))
	recorder := httptest.NewRecorder()
	cfg.handleGetUserBooks(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	var books []testBookResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &books); err != nil {
		t.Fatal(err)
	}
	if len(books) != 2 {
		t.Fatalf("got %d books, want 2", len(books))
	}
}
