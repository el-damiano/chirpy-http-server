package auth

import (
	"errors"
	"fmt"
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
