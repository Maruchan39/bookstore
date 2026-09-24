package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPasswordHash(t *testing.T) {
	plain := "PunchCards9!"
	hash, err := HashPassword(plain)
	if err != nil {
		t.Fatal(err)
	}
	if hash == plain || hash == "" {
		t.Fatalf("expected a password hash, got %q", hash)
	}
	match, err := CheckPasswordHash(plain, hash)
	if err != nil || !match {
		t.Fatalf("matching password failed: match=%v err=%v", match, err)
	}
	match, err = CheckPasswordHash("not-the-password", hash)
	if err != nil || match {
		t.Fatalf("wrong password matched: match=%v err=%v", match, err)
	}
}

func TestJWT(t *testing.T) {
	userID := uuid.New()
	token, err := MakeJWT(userID, "test-signing-key", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ValidateJWT(token, "test-signing-key")
	if err != nil || got != userID {
		t.Fatalf("ValidateJWT() = %v, %v; want %v", got, err, userID)
	}
	if _, err := ValidateJWT(token, "a-different-key"); err == nil {
		t.Fatal("token accepted with the wrong signing key")
	}

	expired, err := MakeJWT(userID, "test-signing-key", -time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateJWT(expired, "test-signing-key"); err == nil {
		t.Fatal("expired token was accepted")
	}
}

func TestGetBearerToken(t *testing.T) {
	cases := []struct {
		name   string
		header string
		want   string
		err    string
	}{
		{"header is absent", "", "", "missing Authorization header"},
		{"uses basic auth", "Basic username:password", "", "malformed Authorization header"},
		{"bearer value is empty", "Bearer ", "", "missing bearer token"},
		{"contains a token", "Bearer token-from-request", "token-from-request", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			headers := http.Header{}
			if tc.header != "" {
				headers.Set("Authorization", tc.header)
			}
			got, err := GetBearerToken(headers)
			gotErr := ""
			if err != nil {
				gotErr = err.Error()
			}
			if got != tc.want || gotErr != tc.err {
				t.Fatalf("GetBearerToken() = %q, %q; want %q, %q", got, gotErr, tc.want, tc.err)
			}
		})
	}
}
