package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	pass, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return "", err
	}
	return string(pass), nil
}

func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return err
	}
	return nil
}

func MakeJWT(userId uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	expirationDate := &jwt.NumericDate{
		Time: time.Now().Add(expiresIn * time.Second),
	}
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    "chirpy",
			Subject:   userId.String(),
			ExpiresAt: expirationDate,
			IssuedAt:  &jwt.NumericDate{Time: time.Now().UTC()},
		})

	tokenSigned, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return tokenSigned, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	type MyCustomClaims struct {
		jwt.RegisteredClaims
	}

	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})

	if err != nil {
		return uuid.UUID{}, err
	} else if claims, ok := token.Claims.(*MyCustomClaims); ok {
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			fmt.Printf("err %v\n", err)
			return uuid.UUID{}, err
		}

		return userID, nil
	} else {
		return uuid.UUID{}, errors.New("unknown claim type, cannot proceed")
	}
}

func BearerToken(headers http.Header) (string, error) {
	headerAuth, ok := headers["Authorization"]
	if !ok {
		return "", errors.New("No authorization header found")
	}

	headerClean := strings.Fields(headerAuth[0])
	if len(headerClean) != 2 {
		return "", fmt.Errorf("Error, too many fields: got %d, want 2. Expected format: Bearer <token>", len(headerClean))
	}
	if headerClean[0] != "Bearer" {
		return "", errors.New("No bearer token found")
	}

	return headerClean[1], nil
}

func APIKey(headers http.Header) (string, error) {
	headerAuth, ok := headers["Authorization"]
	if !ok {
		return "", errors.New("No authorization header found")
	}

	headerClean := strings.Fields(headerAuth[0])
	if len(headerClean) != 2 {
		return "", fmt.Errorf("Error, too many fields: got %d, want 2. Expected format: Bearer <token>", len(headerClean))
	}
	if headerClean[0] != "ApiKey" {
		return "", errors.New("No ApiKey found")
	}

	return headerClean[1], nil
}

func MakeRefreshToken() (string, error) {
	tokenRaw := make([]byte, 32)
	_, err := rand.Read(tokenRaw)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(tokenRaw), nil
}
