package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHash(t *testing.T) {
	const password1 = "one of the passwords of all time fr fr"
	const password2 = "Two Time by crymelt it's a really good animation on youtube"

	hash1, err := HashPassword(password1)
	if err != nil {
		t.Errorf("Error during password hashing: %s", err)
	}

	hash2, err := HashPassword(password2)
	if err != nil {
		t.Errorf("Error during password hashing: %s", err)
	}

	cases := map[string]struct {
		password string
		hash     string
		wantErr  bool
	}{
		"correct password": {
			password: password1,
			hash:     hash1,
			wantErr:  false,
		},
		"incorrect password": {
			password: "woops",
			hash:     hash1,
			wantErr:  true,
		},
		"password doesn't match different hash": {
			password: password1,
			hash:     hash2,
			wantErr:  true,
		},
		"empty password": {
			password: "",
			hash:     hash1,
			wantErr:  true,
		},
		"invalid hash": {
			password: password1,
			hash:     "wooOOOOOooooo00000000000oo",
			wantErr:  true,
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			err = CheckPasswordHash(c.password, c.hash)
			if (err != nil) != c.wantErr {
				t.Errorf("Error during password checking: %s", err)
			}

		})
	}
}

func TestJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "totes secret"

	tokenValid, _ := MakeJWT(userID, tokenSecret, time.Hour)
	tokenExpired, err := MakeJWT(userID, "totes secret", 0)
	if err != nil {
		t.Logf("Error during JWT creation: %v", err)
		t.Fatal()
	}

	_ = tokenValid
	_ = tokenExpired

	cases := map[string]struct {
		expectedID  uuid.UUID
		tokenString string
		tokenSecret string
		wantErr     bool
	}{
		"valid token": {
			expectedID:  userID,
			tokenString: tokenValid,
			tokenSecret: tokenSecret,
			wantErr:     false,
		},
		"expired token": {
			expectedID:  uuid.Nil,
			tokenString: tokenExpired,
			tokenSecret: tokenSecret,
			wantErr:     true,
		},
		"invalid token": {
			expectedID:  uuid.Nil,
			tokenString: "uhhhhhhh",
			tokenSecret: tokenSecret,
			wantErr:     true,
		},
		"empty token": {
			expectedID:  uuid.Nil,
			tokenString: "",
			tokenSecret: tokenSecret,
			wantErr:     true,
		},
		"wrong secret": {
			expectedID:  uuid.Nil,
			tokenString: tokenValid,
			tokenSecret: "BZZZZ wrooooooong",
			wantErr:     true,
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			gotUserID, err := ValidateJWT(c.tokenString, c.tokenSecret)

			if (err != nil) != c.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, c.wantErr)
				return
			}
			if gotUserID != c.expectedID {
				t.Errorf("ValidateJWT() gotUserID = %v, want %v", gotUserID, c.expectedID)
			}
		})
	}
}
