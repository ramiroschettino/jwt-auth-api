package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Endpoints
func NewRouter(handler *APIHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/register", handler.Register)
	r.Post("/login", handler.Login)

	r.Group(func(r chi.Router) {
		r.Use(handler.JWTAuthMiddleware)
		r.Post("/logout", handler.Logout)
		r.Post("/notes", handler.CreateNote)
		r.Get("/notes", handler.GetNotes)
	})

	return r
}

func ListenAndServe(addr string, handler http.Handler) error {
	return http.ListenAndServe(addr, handler)
}
