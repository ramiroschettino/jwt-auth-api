package services

import (
	"testing"
	"time"

	"github.com/ramiroschettino/jwt-auth-api/internal/config"
	"github.com/ramiroschettino/jwt-auth-api/internal/models"
	"github.com/ramiroschettino/jwt-auth-api/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type AuthServiceTestSuite struct {
	suite.Suite
	db          *gorm.DB
	authService *AuthService
}

func (s *AuthServiceTestSuite) SetupTest() {
	var err error
	s.db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	s.NoError(err)

	err = s.db.AutoMigrate(&models.User{}, &models.Note{})
	s.NoError(err)

	cfg := &config.Config{
		JWTSecret:     "test-secret",
		JWTExpiration: 15 * time.Minute,
	}

	userRepo := repositories.NewUserRepository(s.db)
	s.authService = NewAuthService(userRepo, cfg)
}

func TestAuthService(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}

// TestAuthService prueba el registro y login de usuarios
func (s *AuthServiceTestSuite) TestAuthService() {
	t := s.T()

	t.Run("Register", func(t *testing.T) {
		user, err := s.authService.Register("testuser", "testpass", "user")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "user", user.Role)
		assert.NotEmpty(t, user.Password)

		var dbUser models.User
		err = s.db.Where("username = ?", "testuser").First(&dbUser).Error
		assert.NoError(t, err)
		assert.Equal(t, user.ID, dbUser.ID)
	})

	t.Run("Login", func(t *testing.T) {
		_, err := s.authService.Register("testuser2", "testpass", "user")
		assert.NoError(t, err)

		token, err := s.authService.Login("testuser2", "testpass")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		_, err = s.authService.Login("testuser2", "wrongpass")
		assert.Error(t, err)
		assert.Equal(t, "contraseña incorrecta", err.Error())

		_, err = s.authService.Login("nonexistent", "testpass")
		assert.Error(t, err)
		assert.Equal(t, "usuario no encontrado", err.Error())
	})
}
