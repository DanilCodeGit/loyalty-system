package entity

import (
	"errors"

	"github.com/google/uuid"
)

// User is an aggregate for managing users.
type User struct {
	UUID         uuid.UUID
	Login        string
	PasswordHash string
	JWT          string
}

var (
	// ErrUserNotFound is returned when a user is not found.
	ErrUserNotFound = errors.New("user not found")
	// ErrUserExists is returned when a user already exists.
	ErrUserExists = errors.New("user already exists")
	// ErrUserWrongPasswordOrLogin is returned when a user password is wrong.
	ErrUserWrongPasswordOrLogin = errors.New("wrong password")
)

// NewUser returns a new user.
func NewUser(login, passwordHash, jwtToken string, userUUID uuid.UUID) *User {

	user := &User{}

	if userUUID != uuid.Nil {
		user.UUID = userUUID
	} else {
		user.UUID = uuid.New()
	}
	user.Login = login
	user.PasswordHash = passwordHash
	user.JWT = jwtToken

	return user
}
