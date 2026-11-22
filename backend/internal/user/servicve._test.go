package user

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestCreateToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

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
