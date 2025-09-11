# JWT Auth API en Go

Proyecto demo para autenticación con JWT. Incluye registro, login, refresh tokens, rutas protegidas y roles.

## Instalación
1. Clona el repo.
2. `go mod tidy`
3. Copia .env.example a .env y configura.
4. `docker-compose up -d` para DB.
5. `go run cmd/api/main.go`

## Endpoints
- POST /register {username, password, role}
- POST /login {username, password} -> {access_token, refresh_token}
- POST /refresh {refresh_token} -> {access_token}
- POST /notes {title, content} (protegido)
- GET /notes (protegido)

## Tests
`go test ./tests`

## Por qué es impactante
- Clean Architecture.
- JWT con claims y refresh.
- DB persistente con GORM.
- Middleware para auth.
- Listo para producción (agrega más como HTTPS, rate limiting).