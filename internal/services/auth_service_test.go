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
	s.db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	s.NoError(err)

	err = s.db.AutoMigrate(&models.User{}, &models.Note{}, &models.Session{})
	s.NoError(err)

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
		// Verifica que el usuario se registre correctamente y se almacene en la base
		user, err := s.authService.Register("testuser", "testpass", "user")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "user", user.Role)
		assert.NotEmpty(t, user.Password)

		// Confirma que el usuario se guardó en la base de datos
		var dbUser models.User
		err = s.db.Where("username = ?", "testuser").First(&dbUser).Error
		assert.NoError(t, err)
		assert.Equal(t, user.ID, dbUser.ID)
	})

	t.Run("Login", func(t *testing.T) {
		// Verifica login exitoso, login con contraseña incorrecta y usuario inexistente
		_, err := s.authService.Register("testuser2", "testpass", "user")
		assert.NoError(t, err)

		token, err := s.authService.Login("testuser2", "testpass", "test-agent", "127.0.0.1")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Prueba login con contraseña incorrecta
		_, err = s.authService.Login("testuser2", "wrongpass", "test-agent", "127.0.0.1")
		assert.Error(t, err)
		assert.Equal(t, apperrors.ErrInvalidPassword, err)

		// Prueba login con usuario inexistente
		_, err = s.authService.Login("nonexistent", "testpass", "test-agent", "127.0.0.1")
		assert.Error(t, err)
		assert.Equal(t, apperrors.ErrUserNotFound, err)

		// Prueba múltiples logins y control de sesiones activas
		for i := 0; i < maxSessionsPerUser+1; i++ {
			token, err = s.authService.Login("testuser2", "testpass", "test-agent", "127.0.0.1")
			assert.NoError(t, err)
			assert.NotEmpty(t, token)
		}

		// Verifica que solo haya el máximo de sesiones activas permitidas
		var count int64
		result := s.db.Model(&models.Session{}).Where("user_id = ? AND is_active = ?", 1, true).Count(&count)
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(maxSessionsPerUser), count)
	})

	t.Run("JWT Flow", func(t *testing.T) {
		// Prueba el flujo completo: registro, login, validación de token, logout y login nuevo
		user, err := s.authService.Register("jwtuser", "jwtpass", "user")
		assert.NoError(t, err)
		assert.NotNil(t, user)

		// Login exitoso
		token, err := s.authService.Login("jwtuser", "jwtpass", "test-agent", "127.0.0.1")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Validación de token
		userID, role, err := s.authService.ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, userID)
		assert.Equal(t, "user", role)

		// Logout: invalida el token y la sesión
		err = s.authService.Logout(token)
		assert.NoError(t, err)

		// El token ya no es válido después de logout
		_, _, err = s.authService.ValidateToken(token)
		assert.Error(t, err)

		// Login nuevo: genera un nuevo token
		newToken, err := s.authService.Login("jwtuser", "jwtpass", "test-agent", "127.0.0.1")
		assert.NoError(t, err)
		assert.NotEmpty(t, newToken)
		assert.NotEqual(t, token, newToken)

		// El token anterior sigue inválido
		_, _, err = s.authService.ValidateToken(token)
		assert.Error(t, err)

		// El nuevo token es válido
		userID, role, err = s.authService.ValidateToken(newToken)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, userID)
		assert.Equal(t, "user", role)
	})

	t.Run("Role Validation", func(t *testing.T) {
		// Prueba registro y validación de token para diferentes roles
		admin, err := s.authService.Register("adminuser", "adminpass", "admin")
		assert.NoError(t, err)
		assert.NotNil(t, admin)

		regularUser, err := s.authService.Register("regularuser", "userpass", "user")
		assert.NoError(t, err)
		assert.NotNil(t, regularUser)

		// Login y validación de admin
		adminToken, err := s.authService.Login("adminuser", "adminpass", "test-agent", "127.0.0.1")
		assert.NoError(t, err)
		assert.NotEmpty(t, adminToken)

		adminID, role, err := s.authService.ValidateToken(adminToken)
		assert.NoError(t, err)
		assert.Equal(t, admin.ID, adminID)
		assert.Equal(t, "admin", role)

		// Login y validación de usuario regular
		userToken, err := s.authService.Login("regularuser", "userpass", "test-agent", "127.0.0.1")
		assert.NoError(t, err)
		assert.NotEmpty(t, userToken)

		userID, role, err := s.authService.ValidateToken(userToken)
		assert.NoError(t, err)
		assert.Equal(t, regularUser.ID, userID)
		assert.Equal(t, "user", role)
	})
}
