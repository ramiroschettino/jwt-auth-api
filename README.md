# JWT Authentication API

Una API REST moderna para gestión de autenticación y notas personales

## 🚀 Características

- **Autenticación Segura**
  - Sistema completo de JWT con refresh tokens
  - Gestión de sesiones múltiples
  - Hash seguro de contraseñas con bcrypt
  - Control de roles y permisos

- **Gestión de Notas**
  - CRUD completo de notas personales
  - Aislamiento de datos por usuario
  - Validación de entradas

- **Arquitectura Moderna**
  - Clean Architecture
  - Manejo centralizado de errores
  - Tests unitarios extensivos
  - Documentación Swagger

## 🛠️ Tecnologías

- **Go 1.21+** - Backend robusto y eficiente
- **Chi Router** - Enrutamiento HTTP flexible
- **GORM** - ORM potente y developer-friendly
- **PostgreSQL** - Base de datos relacional
- **Docker** - Containerización y despliegue simplificado
- **JWT** - Autenticación stateless
- **Swagger** - Documentación de API

## 📋 Prerrequisitos

- Go 1.21 o superior
- Docker y Docker Compose
- PostgreSQL (incluido en Docker Compose)

## 🚦 Inicio Rápido

1. **Clonar el repositorio**
   ```bash
   git clone https://github.com/ramiroschettino/jwt-auth-api
   cd jwt-auth-api
   ```


2. **Configurar variables de entorno**
   ```bash
   # Crear archivo .env en la raíz del proyecto
   cp .env.example .env
   ```
   Ajusta las variables según tu entorno:
   ```env
   DB_DSN=postgres://user:pass@localhost:5432/jwtdb?sslmode=disable
   JWT_SECRET=tu_secret_key_segura
   JWT_EXPIRATION=15m
   REFRESH_EXPIRATION=24h
   POSTGRES_USER=user
   POSTGRES_PASSWORD=pass
   POSTGRES_DB=jwtdb
   ```


3. **Iniciar servicios con Docker**
   ```bash
   docker-compose up -d
   ```

4. **Instalar dependencias**
   ```bash
   go mod tidy
   ```

5. **Ejecutar la API**
   ```bash
   go run cmd/api/main.go
   ```

El servidor estará disponible en `http://localhost:8080`

## 📡 Endpoints

### Públicos

#### `POST /register`
Registro de nuevos usuarios
```json
{
    "username": "usuario",
    "password": "contraseña",
    "role": "user"
}
```
Respuesta: `201 Created` con datos del usuario


#### `POST /login`
Autenticación de usuarios
```json
{
    "username": "usuario",
    "password": "contraseña"
}
```
Respuesta: `200 OK` con token JWT

### Protegidos
Requieren header: `Authorization: Bearer <token>`

> **Nota sobre roles**: 
> - Rol `admin`: Puede crear y consultar notas
> - Rol `user`: Solo puede consultar notas

#### `POST /notes`
Crear nota personal (solo admin)
```json
{
    "title": "Mi Nota",
    "content": "Contenido de la nota"
}
```
Respuesta: 
- `201 Created` con datos de la nota (para admin)
- `403 Forbidden` si el rol no es admin

#### `GET /notes`
Listar notas del usuario autenticado (disponible para todos los roles)
Respuesta: `200 OK` con array de notas

## 🧪 Tests

Ejecutar suite completa de tests:
```bash
go test ./...
```

## 📚 Documentación

La documentación completa de la API está disponible en:
- Swagger UI: `http://localhost:8080/swagger`
- Docs: `docs/swagger.yaml`

## 🔒 Seguridad

- Todas las contraseñas se hashean con bcrypt
- Tokens JWT con expiración configurable
- Control de sesiones simultáneas
- Validación de entradas en todos los endpoints

## 🤝 Contribuir

1. Fork el proyecto
2. Crea tu Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push al Branch (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request

## 📄 Licencia

Este proyecto está bajo la Licencia MIT - ver el archivo [LICENSE](LICENSE) para más detalles.

## ✨ Autor

Ramiro Schettino - [GitHub](https://github.com/ramiroschettino)

Registrar un usuario:
Método: POST
URL: http://localhost:8080/register
Headers: Content-Type: application/json
Body (raw, JSON): {"username":"testuser","password":"testpass","role":"user"}


Login para obtener un token:
Método: POST
URL: http://localhost:8080/login
Headers: Content-Type: application/json
Body (raw, JSON): {"username":"testuser","password":"testpass"}


Crear una nota:
Método: POST
URL: http://localhost:8080/notes
Headers: Content-Type: application/json, Authorization: Bearer <token>
Body (raw, JSON): {"title":"Mi Nota","content":"Contenido"}


Listar notas:
Método: GET
URL: http://localhost:8080/notes
Headers: Authorization: Bearer <token>



Estructura del Proyecto
jwt-auth-api/
├── .gitignore              # Ignora .env y volúmenes de Docker
├── go.mod                  # Dependencias del módulo Go
├── go.sum                  # Sumas de verificación de dependencias
├── docker-compose.yml      # Configuración de PostgreSQL
├── cmd/
│   └── api/
│       └── main.go         # Punto de entrada de la API
├── internal/
│   ├── config/             # Carga de configuración (.env)
│   ├── models/             # Modelos de datos (User, Note)
│   ├── repositories/       # Operaciones con la base de datos
│   ├── services/           # Lógica de negocio (autenticación, notas)
│   └── handlers/           # Manejadores de HTTP (rutas, controladores)
├── docs/
│   └── swagger.yaml        # Documentación OpenAPI/Swagger
└── tests/                  # Pruebas unitarias y de integración

## Documentación API
La documentación OpenAPI/Swagger está disponible en `/docs/swagger.yaml`

## Mejores Prácticas Implementadas

- Manejo centralizado de errores
- Documentación OpenAPI
- Tests unitarios
- Logs estructurados
- JWT para autenticación
- Clean Architecture
- Docker para la base de datos