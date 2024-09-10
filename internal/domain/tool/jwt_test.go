package tool_test

import (
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	"loyalty/internal/domain/tool"
)

func TestCreateJWT(t *testing.T) {
	userUUID, err := uuid.NewRandom()
	if err != nil {
		t.Fatalf("Failed to create user UUID: %v", err)
	}

	secretKey := "secret"
	jwtString, err := tool.CreateJWT(userUUID, secretKey)
	if err != nil {
		t.Fatalf("Failed to create JWT: %v", err)
	}

	claims := &tool.JWTClaims{}
	_, err = jwt.ParseWithClaims(jwtString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse JWT claims: %v", err)
	}

	if claims.UserID != userUUID {
		t.Errorf("Expected user ID %v, got %v", userUUID, claims.UserID)
	}
}

func TestCheckJWT(t *testing.T) {
	userUUID := uuid.New()

	secretKey := "secret"
	jwtString, err := tool.CreateJWT(userUUID, secretKey)
	if err != nil {
		t.Fatalf("Failed to create JWT: %v", err)
	}

	userID, err := tool.CheckJWT(jwtString, secretKey)
	if err != nil {
		t.Fatalf("Failed to verify JWT: %v", err)
	}

	if userID != userUUID {
		t.Errorf("Expected user ID %v, got %v", userUUID, userID)
	}
}
