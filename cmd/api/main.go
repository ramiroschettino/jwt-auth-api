package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ramiroschettino/jwt-auth-api/internal/config"
	"github.com/ramiroschettino/jwt-auth-api/internal/repositories"
	"github.com/ramiroschettino/jwt-auth-api/internal/services"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Definir tipo propio para claves de contexto
type ctxKey string

const (
	ctxUserID   ctxKey = "user_id"
	ctxUsername ctxKey = "username"
	ctxRole     ctxKey = "role"
)

// APIHandler contiene las dependencias de la API
type APIHandler struct {
	authService *services.AuthService
	noteService *services.NoteService
}

func main() {
	cfg := config.LoadConfig()
	db, err := gorm.Open(postgres.Open(cfg.DBDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Error conectando a la DB: ", err)
	}

	// Inicializar repositorios y servicios
	userRepo := repositories.NewUserRepository(db)
	noteRepo := repositories.NewNoteRepository(db)
	authService := services.NewAuthService(userRepo, cfg)
	noteService := services.NewNoteService(noteRepo)

	handler := &APIHandler{authService: authService, noteService: noteService}

	// Configurar router Chi
	r := chi.NewRouter()
	r.Use(middleware.Logger) // Log de peticiones HTTP

	// Rutas públicas
	r.Post("/register", handler.Register)
	r.Post("/login", handler.Login)

	// Rutas protegidas con JWT
	r.Group(func(r chi.Router) {
		r.Use(handler.JWTAuthMiddleware)
		r.Post("/notes", handler.CreateNote)
		r.Get("/notes", handler.GetNotes)
	})

	log.Println("Servidor escuchando en :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Error iniciando servidor: ", err)
	}
}

// Register maneja el registro de usuarios
func (h *APIHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.authService.Register(req.Username, req.Password, req.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// Login maneja la autenticación
func (h *APIHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// JWTAuthMiddleware verifica el token JWT
func (h *APIHandler) JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}
		if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
			tokenStr = tokenStr[7:]
		}

		claims := jwt.MapClaims{}
		token, err := jwt.NewParser().ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(h.authService.Cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claimsMap, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid claims", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ctxUserID, uint(claimsMap["user_id"].(float64)))
		ctx = context.WithValue(ctx, ctxUsername, claimsMap["username"].(string))
		ctx = context.WithValue(ctx, ctxRole, claimsMap["role"].(string))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CreateNote crea una nota
func (h *APIHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(ctxUserID).(uint)
	note, err := h.noteService.CreateNote(req.Title, req.Content, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(note)
}

// GetNotes lista las notas del usuario
func (h *APIHandler) GetNotes(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ctxUserID).(uint)
	notes, err := h.noteService.GetNotesByUserID(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notes)
}
