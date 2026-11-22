package user

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestCreateToken(t *testing.T) {
	previousSecretKey := secretKey
	secretKey = "test-secret"
	t.Cleanup(func() {
		secretKey = previousSecretKey
	})

	tokenString, err := createToken("alice")
	if err != nil {
		t.Fatal(err)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte("test-secret"), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !token.Valid {
		t.Fatal("expected a valid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["username"] != "alice" {
		t.Fatalf("unexpected claims: %#v", token.Claims)
	}
}

func TestVerifyToken(t *testing.T) {
	previousSecretKey := secretKey
	secretKey = "test-secret"
	t.Cleanup(func() {
		secretKey = previousSecretKey
	})

	tokenString, err := createToken("alice")
	if err != nil {
		t.Fatal(err)
	}

	if err := verifyToken(tokenString); err != nil {
		t.Fatalf("expected token to verify, got %v", err)
	}

	tamperedToken := tokenString[:len(tokenString)-1] + "x"
	if err := verifyToken(tamperedToken); err == nil {
		t.Fatal("expected tampered token verification to fail")
	}
}
