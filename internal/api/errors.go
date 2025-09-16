package api

import (
	"encoding/json"
	"net/http"
)

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewAPIError(code int, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

func WriteError(w http.ResponseWriter, err *APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Code)
	json.NewEncoder(w).Encode(err)
}

var (
	ErrInvalidRequest    = NewAPIError(http.StatusBadRequest, "Invalid request")
	ErrUnauthorized      = NewAPIError(http.StatusUnauthorized, "Unauthorized")
	ErrInvalidToken      = NewAPIError(http.StatusUnauthorized, "Invalid token")
	ErrMissingToken      = NewAPIError(http.StatusUnauthorized, "Missing token")
	ErrInternalServer    = NewAPIError(http.StatusInternalServerError, "Internal server error")
	ErrUserNotFound      = NewAPIError(http.StatusNotFound, "User not found")
	ErrInvalidPassword   = NewAPIError(http.StatusUnauthorized, "Invalid password")
	ErrDuplicateUsername = NewAPIError(http.StatusConflict, "Username already exists")
	ErrInvalidRole       = NewAPIError(http.StatusBadRequest, "Invalid role")
	ErrTokenExpired      = NewAPIError(http.StatusUnauthorized, "Token expired")
)
