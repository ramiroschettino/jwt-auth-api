package services

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/ramiroschettino/jwt-auth-api/internal/config"
	apperrors "github.com/ramiroschettino/jwt-auth-api/internal/errors"
	"github.com/ramiroschettino/jwt-auth-api/internal/models"
	"github.com/ramiroschettino/jwt-auth-api/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// pruebas para el servicio de autenticación
type AuthServiceTestSuite struct {
	suite.Suite
	db          *gorm.DB
	authService *AuthService
}

func (s *AuthServiceTestSuite) SetupTest() {
	var err error
	// Usar cache=shared para que todas las conexiones compartan la misma base de datos en memoria
	s.db, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s.NoError(err)

	// Migrar tablas en orden correcto (sin transacción)
	s.NoError(s.db.AutoMigrate(&models.User{}))
	s.NoError(s.db.AutoMigrate(&models.Note{}))
	s.NoError(s.db.AutoMigrate(&models.Session{}))
	s.NoError(s.db.AutoMigrate(&models.InvalidToken{}))

	// Limpiar datos antes de cada test
	s.db.Exec("DELETE FROM invalid_tokens")
	s.db.Exec("DELETE FROM sessions")
	s.db.Exec("DELETE FROM notes")
	s.db.Exec("DELETE FROM users")

	// Verificar que las tablas existen
	var tables []string
	s.db.Raw("SELECT name FROM sqlite_master WHERE type='table'").Pluck("name", &tables)
	s.Contains(tables, "invalid_tokens")
	s.Contains(tables, "sessions")
	s.Contains(tables, "notes")
	s.Contains(tables, "users")

	// Config de prueba
	cfg := &config.Config{
		JWTSecret:     "test-secret",
		JWTExpiration: 15 * time.Minute,
	}

	userRepo := repositories.NewUserRepository(s.db)
	sessionRepo := repositories.NewSessionRepository(s.db)
	s.authService = NewAuthService(userRepo, sessionRepo, cfg)
}

func TestAuthService(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}

func (s *AuthServiceTestSuite) TestAuthService() {
	t := s.T()

	t.Run("Register", func(t *testing.T) {
		t.Log("Inicio Register test")
		user, err := s.authService.Register("testuser", "testpass", "user")
		t.Logf("Register result: user=%v, err=%v", user, err)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "user", user.Role)
		assert.NotEmpty(t, user.Password)

		// Confirmar que quedó persistido en DB
		var dbUser models.User
		err = s.db.Where("username = ?", "testuser").First(&dbUser).Error
		t.Logf("DB user fetch: dbUser=%v, err=%v", dbUser, err)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, dbUser.ID)
	})

	t.Run("Login", func(t *testing.T) {
		t.Log("Inicio Login test")
		// Limpiar datos antes de comenzar
		s.db.Exec("DELETE FROM invalid_tokens")
		s.db.Exec("DELETE FROM sessions")
		s.db.Exec("DELETE FROM notes")
		s.db.Exec("DELETE FROM users")

		t.Log("Antes de Register testuser2")
		_, err := s.authService.Register("testuser2", "testpass", "user")
		t.Logf("Register testuser2: err=%v", err)
		assert.NoError(t, err)

		t.Log("Antes de Login testuser2")
		token, err := s.authService.Login("testuser2", "testpass", "test-agent", "127.0.0.1")
		t.Logf("Login testuser2: token=%v, err=%v", token, err)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Login con contraseña incorrecta
		t.Log("Antes de Login con contraseña incorrecta")
		_, err = s.authService.Login("testuser2", "wrongpass", "test-agent", "127.0.0.1")
		t.Logf("Login wrongpass: err=%v", err)
		assert.Error(t, err)
		assert.Equal(t, apperrors.ErrInvalidPassword, err)

		// Login con usuario inexistente
		t.Log("Antes de Login usuario inexistente")
		_, err = s.authService.Login("nonexistent", "testpass", "test-agent", "127.0.0.1")
		t.Logf("Login nonexistent: err=%v", err)
		assert.Error(t, err)
		assert.Equal(t, apperrors.ErrUserNotFound, err)

		// Múltiples logins hasta exceder maxSessionsPerUser
		t.Log("Antes de múltiples logins (maxSessionsPerUser+1)")
		for i := 0; i < maxSessionsPerUser+1; i++ {
			token, err = s.authService.Login("testuser2", "testpass", "test-agent", "127.0.0.1")
			t.Logf("Login loop %d: token=%v, err=%v", i, token, err)
			assert.NoError(t, err)
			assert.NotEmpty(t, token)
		}

		// Verificar que solo queden maxSessionsPerUser activas
		t.Log("Antes de verificar sesiones activas")
		var count int64
		// Buscar el usuario testuser2
		var user models.User
		err = s.db.Where("username = ?", "testuser2").First(&user).Error
		t.Logf("Usuario testuser2: id=%v, err=%v", user.ID, err)
		assert.NoError(t, err)
		result := s.db.Model(&models.Session{}).
			Where("user_id = ? AND is_active = ?", user.ID, true).
			Count(&count)
		t.Logf("Sesiones activas: count=%v, err=%v", count, result.Error)
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(maxSessionsPerUser), count)
	})

	t.Run("JWT Flow", func(t *testing.T) {
		// Limpiar sesiones antes de comenzar
		s.db.Exec("DELETE FROM sessions")
		s.db.Exec("DELETE FROM invalid_tokens")

		user, err := s.authService.Register("jwtuser", "jwtpass", "user")
		assert.NoError(t, err)
		assert.NotNil(t, user)

		// Login inicial
		token, err := s.authService.Login("jwtuser", "jwtpass", "test-agent", "127.0.0.1")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Validar token
		userID, role, err := s.authService.ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, userID)
		assert.Equal(t, "user", role)

		// Logout → invalida token
		err = s.authService.Logout(token)
		assert.NoError(t, err)

		_, _, err = s.authService.ValidateToken(token)
		assert.Error(t, err)

		// Nuevo login genera token distinto
		newToken, err := s.authService.Login("jwtuser", "jwtpass", "test-agent", "127.0.0.1")
		assert.NoError(t, err)
		assert.NotEmpty(t, newToken)
		assert.NotEqual(t, token, newToken)

		// Token viejo sigue inválido
		_, _, err = s.authService.ValidateToken(token)
		assert.Error(t, err)

		// Nuevo token válido
		userID, role, err = s.authService.ValidateToken(newToken)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, userID)
		assert.Equal(t, "user", role)
	})

	t.Run("Role Validation", func(t *testing.T) {
		admin, err := s.authService.Register("adminuser", "adminpass", "admin")
		assert.NoError(t, err)

		regularUser, err := s.authService.Register("regularuser", "userpass", "user")
		assert.NoError(t, err)

		// Admin
		adminToken, err := s.authService.Login("adminuser", "adminpass", "test-agent", "127.0.0.1")
		assert.NoError(t, err)
		adminID, role, err := s.authService.ValidateToken(adminToken)
		assert.NoError(t, err)
		assert.Equal(t, admin.ID, adminID)
		assert.Equal(t, "admin", role)

		// Usuario normal
		userToken, err := s.authService.Login("regularuser", "userpass", "test-agent", "127.0.0.1")
		assert.NoError(t, err)
		userID, role, err := s.authService.ValidateToken(userToken)
		assert.NoError(t, err)
		assert.Equal(t, regularUser.ID, userID)
		assert.Equal(t, "user", role)
	})
}
