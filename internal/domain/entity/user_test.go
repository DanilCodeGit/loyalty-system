package entity_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"loyalty/internal/domain/entity"
)

func TestNewUser(t *testing.T) {
	t.Run("Should return a new user", func(t *testing.T) {
		login := "test-user"
		passwordHash := "test-password"
		jwtToken := "test-jwt"
		userUUID := uuid.New()

		user := entity.NewUser(login, passwordHash, jwtToken, userUUID)

		assert.Equal(t, userUUID, user.UUID)
		assert.Equal(t, login, user.Login)
		assert.Equal(t, passwordHash, user.PasswordHash)
		assert.Equal(t, jwtToken, user.JWT)
	})
	t.Run("Should return a new empty user with generated uuid", func(t *testing.T) {

		user := entity.NewUser("", "", "", uuid.Nil)

		assert.IsType(t, uuid.New(), user.UUID)
	})
}
