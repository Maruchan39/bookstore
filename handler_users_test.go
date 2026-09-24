package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Maruchan39/bookstore/internal/auth"
)

func TestHandleCreateUser(t *testing.T) {
	cfg := newTestConfig(t)
	body := map[string]string{
		"firstName": "Sana",
		"lastName":  "Voss",
		"username":  "sana@analytical.test",
		"password":  "PunchCards9!",
	}
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	cfg.handleCreateUser(recorder, httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(data)))
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Token == "" {
		t.Fatal("signup returned an empty token")
	}

	var user User
	if err := cfg.db.Where("username = ?", body["username"]).First(&user).Error; err != nil {
		t.Fatal(err)
	}
	if user.HashedPassword == body["password"] {
		t.Fatal("password was stored in plaintext")
	}
	match, err := auth.CheckPasswordHash(body["password"], user.HashedPassword)
	if err != nil || !match {
		t.Fatalf("stored password does not verify: match=%v err=%v", match, err)
	}
}

func TestHandleCreateUserRejectsInvalidDetails(t *testing.T) {
	cfg := newTestConfig(t)
	cases := []struct {
		name      string
		firstName string
		username  string
		password  string
		want      string
	}{
		{"first name is too short", "A", "eva@books.test", "PunchCards9!", "firstName must have 2 or more non-whitespace characters"},
		{"username is not an email", "Ieva", "not an email", "PunchCards9!", "username must be a valid email address"},
		{"password lacks the required mix", "Ieva", "eva@books.test", "plainpassword", "password must be at least 8 characters long and contain a lowercase letter, an uppercase letter, a number and a special character"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(map[string]string{
				"firstName": tc.firstName,
				"lastName":  "Stone",
				"username":  tc.username,
				"password":  tc.password,
			})
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()
			cfg.handleCreateUser(recorder, httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(data)))
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want bad request", recorder.Code)
			}
			var message testErrorResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &message); err != nil {
				t.Fatal(err)
			}
			if message.Error != tc.want {
				t.Fatalf("error = %q, want %q", message.Error, tc.want)
			}
		})
	}
}

func TestHandleCreateUserRejectsDuplicateUsername(t *testing.T) {
	cfg := newTestConfig(t)
	body := map[string]string{
		"firstName": "Jane",
		"lastName":  "Stone",
		"username":  "same@knygos.test",
		"password":  "PunchCards9!",
	}
	data, _ := json.Marshal(body)
	first := httptest.NewRecorder()
	cfg.handleCreateUser(first, httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(data)))
	if first.Code != http.StatusCreated {
		t.Fatalf("first signup status = %d", first.Code)
	}
	data, _ = json.Marshal(body)
	second := httptest.NewRecorder()
	cfg.handleCreateUser(second, httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(data)))
	if second.Code != http.StatusConflict {
		t.Fatalf("second signup status = %d, want conflict", second.Code)
	}
}

func TestHandleLogin(t *testing.T) {
	cfg := newTestConfig(t)
	user := User{
		Username:       "jane@quietcorner.test",
		FirstName:      "Jane",
		LastName:       "Stone",
		HashedPassword: hashPassword(t, "PunchCards9!"),
	}
	if err := cfg.db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name     string
		username string
		password string
		want     int
	}{
		{"correct credentials", user.Username, "PunchCards9!", http.StatusOK},
		{"no such account", "nothing@books.test", "PunchCards9!", http.StatusUnauthorized},
		{"wrong password", user.Username, "NotThePassword4!", http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(map[string]string{"username": tc.username, "password": tc.password})
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()
			cfg.handleLogin(recorder, httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(data)))
			if recorder.Code != tc.want {
				t.Fatalf("status = %d, want %d", recorder.Code, tc.want)
			}
		})
	}
}
