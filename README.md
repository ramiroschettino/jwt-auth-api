
# JWT Auth API

API REST para autenticación de usuarios y gestión de notas personales. Pensada para ser segura, clara y fácil de mantener.

## Características principales

- Autenticación con JWT y control de sesiones
- Roles: `admin` y `user` (admin puede crear notas, user solo consultar)
- CRUD de notas personales
- Arquitectura limpia y modular
- Manejo centralizado de errores
- Documentación Swagger interactiva
- Tests unitarios

## Tecnologías

- Go 1.21+
- Chi Router
- GORM
- PostgreSQL
- Docker
- JWT
- Swagger

## Requisitos

- Go 1.21 o superior
- Docker y Docker Compose

## Cómo empezar


1. Clona el repositorio:
   ```cmd
   git clone https://github.com/ramiroschettino/jwt-auth-api
   cd jwt-auth-api
   ```

2. Configura las variables de entorno:
   ```cmd
   copy .env.example .env
   rem Edita .env según tu entorno
   ```

3. Inicia la base de datos:
   ```cmd
   docker-compose up -d
   ```

4. Instala dependencias:
   ```cmd
   go mod tidy
   ```

5. Ejecuta la API:
   ```cmd
   go run cmd\api\main.go
   ```

El servidor estará disponible en [http://localhost:8080](http://localhost:8080)

## Endpoints principales

### Públicos

- `POST /register` — Registro de usuario
- `POST /login` — Login y obtención de token JWT

### Protegidos (requieren `Authorization: Bearer <token>`)

- `POST /notes` — Crear nota (solo admin)
- `GET /notes` — Listar notas del usuario
- `POST /logout` — Cerrar sesión

**Roles:**
- `admin`: puede crear y consultar notas
- `user`: solo consultar

## Ejemplos de uso rápido

**Registrar usuario**
```http
POST /register
Content-Type: application/json

{
  "username": "testuser",
  "password": "testpass",
  "role": "user"
}
```

**Login**
```http
POST /login
Content-Type: application/json

{
  "username": "testuser",
  "password": "testpass"
}
```

**Crear nota (admin)**
```http
POST /notes
Content-Type: application/json
Authorization: Bearer <token>

{
  "title": "Mi Nota",
  "content": "Contenido"
}
```

**Listar notas**
```http
GET /notes
Authorization: Bearer <token>
```


## Tests

Los tests unitarios están en la carpeta `internal/services`.

Para ejecutarlos desde Windows:
```cmd
go test ./internal/services
```
o para ver detalles:
```cmd
go test -v ./internal/services
```

## Documentación

- Swagger UI: [http://localhost:8080/swagger](http://localhost:8080/swagger)
- OpenAPI: `docs/swagger.yaml`

## Seguridad

- Contraseñas hasheadas con bcrypt
- Expiración configurable de tokens
- Control de sesiones simultáneas
- Validación de entradas

## Estructura del proyecto

```
jwt-auth-api/
├── cmd/api/main.go         # Punto de entrada
├── internal/config/        # Configuración
├── internal/models/        # Modelos de datos
├── internal/repositories/  # Acceso a datos
├── internal/services/      # Lógica de negocio
├── internal/api/           # Rutas y controladores
├── docs/swagger.yaml       # Documentación OpenAPI
├── docker-compose.yml      # Base de datos
└── tests/                  # Pruebas
```

## Contribuir

¿Te gustaría mejorar el proyecto? ¡Bienvenido! Haz un fork, crea tu rama y abre un PR.

## Autor

Ramiro Schettino — [GitHub](https://github.com/ramiroschettino)