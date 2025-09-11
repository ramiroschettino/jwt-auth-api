API de Autenticación JWT
  Una API RESTful desarrollada con Go, Chi, GORM y PostgreSQL para autenticación de usuarios (usando JWT) y gestión de notas privadas.
Características

Registro y login de usuarios con autenticación basada en JWT
Creación y listado de notas privadas por usuario
Hash seguro de contraseñas con bcrypt
Base de datos PostgreSQL gestionada con Docker
Arquitectura limpia (modelos, repositorios, servicios, manejadores)

Requisitos

Go (versión 1.21 o superior)
Docker Desktop para Windows
Postman para probar los endpoints

Configuración

Clonar el repositorio:
git clone https://github.com/ramiroschettino/jwt-auth-api
cd jwt-auth-api


Crear el archivo .env:En la carpeta raíz (C:\proyectos\jwt-auth-api), crea un archivo .env con:
DB_DSN=postgres://user:pass@localhost:5432/jwtdb?sslmode=disable
JWT_SECRET=9e8d7c6b5a4f3e2d1c0b9a8f7e6d5c4b3a2f1e0d9c8b7a6f5e4d3c2b1a0f9e
JWT_EXPIRATION=15m
REFRESH_EXPIRATION=24h
POSTGRES_USER=user
POSTGRES_PASSWORD=pass
POSTGRES_DB=jwtdb


Iniciar PostgreSQL con Docker:Abre un símbolo del sistema como administrador y ejecuta:
docker-compose up -d

Verifica que el contenedor esté corriendo:
docker ps


Instalar dependencias de Go:
go mod tidy


Ejecutar la API:
go run cmd\api\main.go

El servidor estará disponible en http://localhost:8080.


Endpoints de la API
Endpoints públicos

POST /register
Descripción: Registra un nuevo usuario.
Body (JSON):{
    "username": "testuser",
    "password": "testpass",
    "role": "user"
}


Respuesta: Código 201 Created con datos del usuario (ID, username, role, timestamps).


POST /login
Descripción: Autentica un usuario y devuelve un token JWT.
Body (JSON):{
    "username": "testuser",
    "password": "testpass"
}


Respuesta: Código 200 OK con {"token": "jwt_token"}.



Endpoints protegidos (requieren Authorization: Bearer <token>)

POST /notes
Descripción: Crea una nota para el usuario autenticado.
Headers: Authorization: Bearer <token>
Body (JSON):{
    "title": "Mi Nota",
    "content": "Contenido"
}


Respuesta: Código 201 Created con datos de la nota.


GET /notes
Descripción: Lista las notas del usuario autenticado.
Headers: Authorization: Bearer <token>
Respuesta: Código 200 OK con un array de notas.



Probar con Postman

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