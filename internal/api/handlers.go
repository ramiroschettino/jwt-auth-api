package api

import (
	"context"
	"encoding/json"
	"net/http"

	apperrors "github.com/ramiroschettino/jwt-auth-api/internal/errors"
	"github.com/ramiroschettino/jwt-auth-api/internal/services"
)

type ctxKey string

const (
	ctxUserID   ctxKey = "user_id"
	ctxUsername ctxKey = "username"
	ctxRole     ctxKey = "role"
)

type APIHandler struct {
	AuthService *services.AuthService
	NoteService *services.NoteService
}

func NewAPIHandler(auth *services.AuthService, note *services.NoteService) *APIHandler {
	return &APIHandler{AuthService: auth, NoteService: note}
}

func (h *APIHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, MapError(apperrors.ErrInvalidUser))
		return
	}

	if req.Role != "user" && req.Role != "admin" {
		WriteError(w, MapError(apperrors.ErrInvalidRole))
		return
	}

	user, err := h.AuthService.Register(req.Username, req.Password, req.Role)
	if err != nil {
		WriteError(w, MapError(err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *APIHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, MapError(apperrors.ErrInvalidUser))
		return
	}

	if req.Username == "" || req.Password == "" {
		WriteError(w, MapError(apperrors.ErrInvalidUser))
		return
	}

	userAgent := r.Header.Get("User-Agent")
	ip := r.RemoteAddr
	if fwdIP := r.Header.Get("X-Forwarded-For"); fwdIP != "" {
		ip = fwdIP
	}

	token, err := h.AuthService.Login(req.Username, req.Password, userAgent, ip)
	if err != nil {
		WriteError(w, MapError(err))
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *APIHandler) JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			WriteError(w, MapError(apperrors.ErrTokenMissing))
			return
		}
		if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
			tokenStr = tokenStr[7:]
		}

		userID, role, err := h.AuthService.ValidateToken(tokenStr)
		if err != nil {
			WriteError(w, MapError(err))
			return
		}

		ctx := context.WithValue(r.Context(), ctxUserID, userID)
		ctx = context.WithValue(ctx, ctxRole, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *APIHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, MapError(apperrors.ErrInvalidUser))
		return
	}

	roleVal := r.Context().Value(ctxRole)
	role, ok := roleVal.(string)
	if !ok || role != "admin" {
		WriteError(w, NewAPIError(http.StatusForbidden, "solo los administradores pueden crear notas"))
		return
	}

	userIDVal := r.Context().Value(ctxUserID)
	userID, ok := userIDVal.(uint)
	if !ok {
		WriteError(w, MapError(apperrors.ErrUnauthorized))
		return
	}
	note, err := h.NoteService.CreateNote(req.Title, req.Content, userID)
	if err != nil {
		WriteError(w, MapError(err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(note)
}

func (h *APIHandler) GetNotes(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value(ctxUserID)
	userID, ok := userIDVal.(uint)
	if !ok {
		WriteError(w, MapError(apperrors.ErrUnauthorized))
		return
	}
	notes, err := h.NoteService.GetNotesByUserID(userID)
	if err != nil {
		WriteError(w, MapError(err))
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notes)
}

func (h *APIHandler) Logout(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.Header.Get("Authorization")
	if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
		tokenStr = tokenStr[7:]
	}

	if err := h.AuthService.Logout(tokenStr); err != nil {
		WriteError(w, MapError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Successfully logged out"})
}
