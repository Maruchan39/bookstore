package main

import "testing"

func TestIsValidEmail(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  bool
	}{
		{"ordinary address", "mira@quietcorner.test", true},
		{"plus sign is okay", "mira+books@quietcorner.co.uk", true},
		{"local address without a dot", "mira@quietcorner", true},
		{"space in local part", "mira stone@quietcorner.test", false},
		{"trailing space", "mira@quietcorner.test ", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isValidEmail(tc.value); got != tc.want {
				t.Fatalf("isValidEmail(%q) = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}

func TestIsStrongPassword(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  bool
	}{
		{"has every required character type", "PunchCards9!", true},
		{"too short", "Short7!", false},
		{"no uppercase letter", "punchcards9!", false},
		{"no lowercase letter", "PUNCHCARDS9!", false},
		{"no number", "PunchCards!", false},
		{"no punctuation", "PunchCards9", false},
		{"accented lowercase letter still counts", "PünchCards9!", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isStrongPassword(tc.value); got != tc.want {
				t.Fatalf("isStrongPassword(%q) = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}
