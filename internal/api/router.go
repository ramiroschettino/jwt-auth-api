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

	// Swagger UI embebido en /swagger
	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Swagger UI</title>
	<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css" />
</head>
<body>
	<div id="swagger-ui"></div>
	<script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
	<script>
		window.onload = function() {
			SwaggerUIBundle({
				url: '/swagger.yaml',
				dom_id: '#swagger-ui',
			});
		};
	</script>
</body>
</html>
							 `))
	})

	// Servir el archivo swagger.yaml en /swagger.yaml
	r.Get("/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/swagger.yaml")
	})

	return r
}

func ListenAndServe(addr string, handler http.Handler) error {
	return http.ListenAndServe(addr, handler)
}
