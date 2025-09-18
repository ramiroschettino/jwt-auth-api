package api

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/ramiroschettino/jwt-auth-api/internal/errors"
)

// APIError represents an HTTP error response
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewAPIError creates a new API error with the given status code and message
func NewAPIError(code int, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

// WriteError writes an error response to the http.ResponseWriter
func WriteError(w http.ResponseWriter, err *APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Code)
	json.NewEncoder(w).Encode(err)
}

// MapError maps domain errors to API errors
func MapError(err error) *APIError {
	switch {
	case apperrors.IsAuthError(err):
		return NewAPIError(http.StatusUnauthorized, err.Error())
	case apperrors.IsTokenError(err):
		return NewAPIError(http.StatusUnauthorized, err.Error())
	default:
		return ErrInternalServer
	}
}

var (
	// Request validation errors
	ErrInvalidRequest = NewAPIError(http.StatusBadRequest, "Invalid request")
	ErrInvalidRole    = NewAPIError(http.StatusBadRequest, "Invalid role")

	// Authentication errors
	ErrUnauthorized      = NewAPIError(http.StatusUnauthorized, "Unauthorized")
	ErrInvalidToken      = NewAPIError(http.StatusUnauthorized, "Invalid token")
	ErrMissingToken      = NewAPIError(http.StatusUnauthorized, "Missing token")
	ErrTokenExpired      = NewAPIError(http.StatusUnauthorized, "Token expired")
	ErrInvalidPassword   = NewAPIError(http.StatusUnauthorized, "Invalid password")
	ErrDuplicateUsername = NewAPIError(http.StatusConflict, "Username already exists")
	ErrUserNotFound      = NewAPIError(http.StatusNotFound, "User not found")

	// Server errors
	ErrInternalServer = NewAPIError(http.StatusInternalServerError, "Internal server error")
)
