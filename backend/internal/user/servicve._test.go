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

func TestHashAndSalt(t *testing.T) {
	password := []byte("correct-horse-battery-staple")
	hashedPassword, err := hashAndSalt(password)
	if err != nil {
		t.Fatal(err)
	}

	if hashedPassword == "" {
		t.Fatal("expected a password hash")
	}
	if hashedPassword == string(password) {
		t.Fatal("expected the hash to differ from the password")
	}

	if matched := comparePasswords(hashedPassword, password); !matched {
		t.Fatalf("expected password to match, got matched=%v", matched)
	}
}

func TestComparePasswords(t *testing.T) {
	hashedPassword, err := hashAndSalt([]byte("correct-password"))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		hashed      string
		password    []byte
		wantMatched bool
		wantErr     error
	}{
		{
			name:        "matching password",
			hashed:      hashedPassword,
			password:    []byte("correct-password"),
			wantMatched: true,
		},
		{
			name:     "incorrect password",
			hashed:   hashedPassword,
			password: []byte("wrong-password"),
			wantErr:  ErrInvalidPassword,
		},
		{
			name:     "invalid hash",
			hashed:   "not-a-bcrypt-hash",
			password: []byte("correct-password"),
			wantErr:  ErrInvalidPassword,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			matched := comparePasswords(test.hashed, test.password)
			if matched != test.wantMatched {
				t.Errorf("matched = %v, want %v", matched, test.wantMatched)
			}
		})
	}
}
