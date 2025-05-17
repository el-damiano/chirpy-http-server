package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestHash(t *testing.T) {
	cases := map[string]struct {
		input    string
		expected string
	}{
		"empty string": {
			input:    "",
			expected: "",
		},
		"password: foo": {
			input:    "foo",
			expected: "foo",
		},
		"password: super strong password": {
			input:    "one of the passwords of all time fr fr",
			expected: "one of the passwords of all time fr fr",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {

			actual, err := HashPassword(c.input)
			if err != nil {
				t.Errorf("Error during password hashing: %s", err)
			}

			err = CheckPasswordHash(actual, c.expected)
			if err != nil {
				t.Errorf("Error during password checking: %s", err)
			}

		})
	}
}

func TestHashWrong(t *testing.T) {
	passwordSet := "foo"
	passwordWrong := "bar"

	passwordHashed, err := HashPassword(passwordSet)
	if err != nil {
		t.Errorf("Error during password hashing: %s", err)
	}

	err = CheckPasswordHash(passwordHashed, passwordWrong)
	if err == nil {
		t.Errorf("Error during password checking: %s. Expected wrong password.", err)
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
