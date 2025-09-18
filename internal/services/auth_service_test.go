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

		// Test successful login
		token, err := s.authService.Login("testuser2", "testpass", "test-agent", "127.0.0.1")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Test wrong password
		_, err = s.authService.Login("testuser2", "wrongpass", "test-agent", "127.0.0.1")
		assert.Error(t, err)
		assert.Equal(t, apperrors.ErrInvalidPassword, err)

		// Test nonexistent user
		_, err = s.authService.Login("nonexistent", "testpass", "test-agent", "127.0.0.1")
		assert.Error(t, err)
		assert.Equal(t, apperrors.ErrUserNotFound, err)

		// Test multiple sessions
		for i := 0; i < maxSessionsPerUser+1; i++ {
			token, err = s.authService.Login("testuser2", "testpass", "test-agent", "127.0.0.1")
			assert.NoError(t, err)
			assert.NotEmpty(t, token)
		}

		// Verify that only maxSessionsPerUser sessions exist
		var count int64
		result := s.db.Model(&models.Session{}).Where("user_id = ? AND is_active = ?", 1, true).Count(&count)
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(maxSessionsPerUser), count)
	})
}
