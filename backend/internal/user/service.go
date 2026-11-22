package user

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrSecretKeyNotSet = errors.New("The secret key for generating the JWT is not set")
	ErrInvalidKey      = errors.New("Failed to login, the username or password provided is incorrect")
)

var secretKey = os.Getenv("JWT_SECRET")

func createToken(username string) (string, error) {
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

func verifyToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return ErrInvalidKey
	}

	return nil
}
