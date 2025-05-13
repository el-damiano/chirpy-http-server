package auth

import (
	"fmt"
	"testing"
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
