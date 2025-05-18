package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
	cases := map[string]struct {
		expectedID  uuid.UUID
		tokenSecret string
		expiresIn   time.Duration
		expected    string
	}{
		"valid case": {
			expectedID:  uuid.MustParse("30fc2344-6f45-46dc-acc2-f6c5d2a26920"),
			tokenSecret: "totes secret",
			expiresIn:   5,
			expected:    "yes",
		},
		"empty token": {
			expectedID:  uuid.MustParse("30fc2344-6f45-46dc-acc2-f6c5d2a26920"),
			tokenSecret: "",
			expiresIn:   5,
			expected:    "yes",
		},
		"zero user id": {
			expectedID:  uuid.MustParse("00000000-0000-0000-0000-000000000000"),
			tokenSecret: "totes secret",
			expiresIn:   5,
			expected:    "yes",
		},
		"expired": {
			expectedID:  uuid.MustParse("30fc2344-6f45-46dc-acc2-f6c5d2a26920"),
			tokenSecret: "",
			expiresIn:   0,
			expected:    fmt.Sprintf("%s: %s", jwt.ErrTokenInvalidClaims, jwt.ErrTokenExpired),
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			token, err := MakeJWT(c.expectedID, c.tokenSecret, c.expiresIn)
			if err != nil {
				t.Logf("Failed making JWT: %v", err)
				t.Fail()
			}

			userID, err := ValidateJWT(token, c.tokenSecret)
			if err != nil {
				if c.expiresIn == 0 && err.Error() == c.expected {
					return
				}
				t.Logf("Failed validating JWT: %v", err)
				t.Fail()
			}

			if c.expectedID != userID {
				t.Logf("Failed comparing the user ID: %v", err)
				t.Fail()
			}

		})
	}

}
