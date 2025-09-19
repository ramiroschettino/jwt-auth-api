package api

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/ramiroschettino/jwt-auth-api/internal/errors"
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

func MapError(err error) *APIError {
	switch {
	case apperrors.IsAuthError(err):
		return NewAPIError(http.StatusUnauthorized, err.Error())
	case apperrors.IsTokenError(err):
		return NewAPIError(http.StatusUnauthorized, err.Error())
	default:
		return NewAPIError(http.StatusInternalServerError, "Internal server error")
	}
}
