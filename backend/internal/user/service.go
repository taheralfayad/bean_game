package user

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrSecretKeyNotSet = errors.New("The secret key for generating the JWT is not set")
)

func createToken(username string) (string, error) {
	secretKey := os.Getenv("JWT_SECRET")

	if secretKey == "" {
		return "", ErrSecretKeyNotSet
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"exp":      time.Now().Add(15 * time.Minute).Unix(),
		})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
