package errors

import (
	"errors"
	"fmt"
)

var (
	// Authentication errors
	ErrUserExists      = errors.New("username already exists")
	ErrInvalidUser     = errors.New("invalid username or password")
	ErrInvalidPassword = errors.New("invalid password")
	ErrUserNotFound    = errors.New("user not found")

	// JWT token errors
	ErrTokenInvalid     = errors.New("token is invalid")
	ErrTokenExpired     = errors.New("token has expired")
	ErrTokenBlacklisted = errors.New("token has been invalidated")
	ErrTokenMissing     = errors.New("token is missing")

	// Authorization errors
	ErrUnauthorized = errors.New("unauthorized access")
	ErrInvalidRole  = errors.New("invalid role")
)

func WrapError(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}

func IsTokenError(err error) bool {
	return errors.Is(err, ErrTokenInvalid) ||
		errors.Is(err, ErrTokenExpired) ||
		errors.Is(err, ErrTokenBlacklisted) ||
		errors.Is(err, ErrTokenMissing)
}

func IsAuthError(err error) bool {
	return errors.Is(err, ErrUserExists) ||
		errors.Is(err, ErrInvalidUser) ||
		errors.Is(err, ErrInvalidPassword) ||
		errors.Is(err, ErrUserNotFound)
}
