package user

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrSecretKeyNotSet = errors.New("The secret key for generating the JWT is not set")
	ErrInvalidKey      = errors.New("The request is not authorized")
	ErrInvalidPassword = errors.New("The username or password you entered is invalid")
)

var secretKey = os.Getenv("JWT_SECRET")

type userStore interface {
	GetUser(ctx context.Context, username string) (*User, string, error)
	InsertUser(ctx context.Context, username, hashedPassword string, role Role) (*User, error)
}

type Service struct {
	store userStore
}

func NewService(store userStore) *Service {
	return &Service{store: store}
}

func (s *Service) Get(ctx context.Context, username string, password []byte) (*User, error) {
	user, hashedPassword, err := s.store.GetUser(ctx, username)
	if err != nil || !comparePasswords(hashedPassword, password) {
		return nil, ErrInvalidPassword
	}

	token, err := createToken(username)
	if err != nil {
		return nil, err
	}
	user.Token = token
	return user, nil
}

func (s *Service) Create(ctx context.Context, username string, password []byte, role Role) (*User, error) {
	hashedPassword, err := hashAndSalt(password)
	if err != nil {
		return nil, err
	}
	token, err := createToken(username)
	if err != nil {
		return nil, err
	}

	user, err := s.store.InsertUser(ctx, username, hashedPassword, role)
	if err != nil {
		return nil, err
	}
	user.Token = token
	return user, nil
}

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

func hashAndSalt(pwd []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func comparePasswords(hashedPwd string, plainPwd []byte) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	return err == nil
}
