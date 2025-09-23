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

## Cómo empezar (paso a paso)

### 1. Clona el repositorio
```cmd
git clone https://github.com/ramiroschettino/jwt-auth-api
cd jwt-auth-api
```

### 2. Crea tu archivo de configuración

**Crea un archivo `.env` en la raíz del proyecto con exactamente este contenido:**

```env
# === Configuración de PostgreSQL ===
POSTGRES_USER=jwt_user
POSTGRES_PASSWORD=jwt_pass_2025
POSTGRES_DB=jwt_auth_db

# === Configuración de conexión a la base de datos ===
DB_HOST=127.0.0.1
DB_PORT=5433
DB_USER=jwt_user
DB_PASSWORD=jwt_pass_2025
DB_NAME=jwt_auth_db
DB_SSLMODE=disable

# === Configuración JWT ===
JWT_SECRET=kj82hfs9d8fhs7df98hsdf78hsdf
JWT_EXPIRATION=15m
REFRESH_SECRET=z9x8c7v6b5n4m3a2q1w0r9t8y7u6i5
REFRESH_EXPIRATION=24h

# === Configuración del servidor ===
PORT=8080
ENV=development
```

> ⚠️ **Importante**: 
> - Cambia `JWT_SECRET` y `REFRESH_SECRET` por valores únicos en producción
> - El puerto de la base de datos es `5433` (no 5432) para evitar conflictos

### 3. Instala dependencias
```cmd
go mod tidy
```

### 4. Inicia la base de datos
```cmd
docker-compose up -d
```

**Verifica que PostgreSQL esté listo:**
```cmd
docker-compose exec db pg_isready -U jwt_user -d jwt_auth_db
```

Deberías ver: `accepting connections`

### 5. Ejecuta la API
```cmd
go run cmd/api/main.go
```

**Si todo está bien, verás:**
```
Servidor escuchando en :8080
```

### 6. Verifica que funciona
Abre tu navegador en [http://localhost:8080](http://localhost:8080) o usa curl:

```cmd
curl http://localhost:8080/health
```

## Solución de problemas comunes

### Error: "falta la variable de entorno requerida"
- **Causa**: El archivo `.env` no se encuentra o está mal ubicado
- **Solución**: Asegúrate de que `.env` esté en la carpeta raíz del proyecto (mismo nivel que `go.mod`)

### Error: "failed to connect to database" 
- **Causa**: PostgreSQL no está listo o hay conflicto de puertos
- **Solución**:
  1. Espera unos segundos más después de `docker-compose up -d`
  2. Verifica que no tengas otro PostgreSQL corriendo en puerto 5432
  3. Usa `docker-compose down -v` y luego `docker-compose up -d` para reiniciar limpio

### Error: "port already in use"
- **Causa**: Ya tienes algo corriendo en puerto 8080
- **Solución**: Cambia `PORT=8080` por `PORT=8081` en tu `.env`

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

### 1. Registrar un usuario admin
```http
POST http://localhost:8080/register
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123",
  "role": "admin"
}
```

### 2. Login
```http
POST http://localhost:8080/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

**Respuesta:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "admin",
    "role": "admin"
  }
}
```

### 3. Crear nota (usando el token del paso anterior)
```http
POST http://localhost:8080/notes
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "title": "Mi Primera Nota",
  "content": "Este es el contenido de mi nota"
}
```

### 4. Listar notas
```http
GET http://localhost:8080/notes
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## Tests

Para ejecutar los tests unitarios:

```cmd
go test ./internal/services -v
```

Para tests con coverage:
```cmd
go test ./internal/services -cover
```

## Documentación

Una vez que tengas el servidor corriendo:

- **Swagger UI**: [http://localhost:8080/swagger](http://localhost:8080/swagger)
- **OpenAPI**: Revisa `docs/swagger.yaml`

## Para producción

### Variables críticas que debes cambiar:

```env
# Genera secretos únicos y seguros
JWT_SECRET=tu-secreto-super-seguro-aqui
REFRESH_SECRET=otro-secreto-diferente-aqui

# Usa una base de datos dedicada
DB_HOST=tu-servidor-postgres.com
DB_PASSWORD=contraseña-fuerte

# Cambia a producción
ENV=production
```

### Recomendaciones:
- Usa HTTPS en producción
- Configura un reverse proxy (nginx, Traefik)
- Implementa rate limiting
- Monitorea logs y métricas

## Estructura del proyecto

```
jwt-auth-api/
├── cmd/api/main.go         # Punto de entrada
├── internal/
│   ├── config/             # Configuración y .env
│   ├── models/             # Modelos de datos (User, Note, etc.)
│   ├── repositories/       # Acceso a base de datos
│   ├── services/           # Lógica de negocio
│   └── api/                # Rutas, handlers, middleware
├── docs/swagger.yaml       # Documentación OpenAPI
├── docker-compose.yml      # PostgreSQL
├── .env                    # Tu configuración (crear este archivo)
└── README.md              # Este archivo
```

## Comandos útiles

```cmd
# Reiniciar base de datos limpia
docker-compose down -v && docker-compose up -d

# Ver logs de PostgreSQL
docker-compose logs db

# Conectarse a PostgreSQL directamente
docker-compose exec db psql -U jwt_user -d jwt_auth_db

# Ejecutar con logs detallados
go run cmd/api/main.go -v

# Build para producción
go build -o jwt-auth-api cmd/api/main.go
```

## Contribuir

¿Te gustaría mejorar el proyecto? ¡Bienvenido! 

1. Haz un fork del proyecto
2. Crea tu rama: `git checkout -b feature/nueva-caracteristica`
3. Commit tus cambios: `git commit -am 'Agrega nueva característica'`
4. Push a la rama: `git push origin feature/nueva-caracteristica`
5. Abre un Pull Request

## Problemas conocidos

Si encuentras algún bug o tienes sugerencias, por favor abre un [issue en GitHub](https://github.com/ramiroschettino/jwt-auth-api/issues).

## Autor

**Ramiro Schettino** — [GitHub](https://github.com/ramiroschettino)

---

¿Te fue útil este proyecto? ¡Dale una ⭐ en GitHub!