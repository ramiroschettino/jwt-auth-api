package services

import (
	"testing"
	"time"

	"github.com/ramiroschettino/jwt-auth-api/internal/config"
	"github.com/ramiroschettino/jwt-auth-api/internal/models"
	"github.com/ramiroschettino/jwt-auth-api/internal/repositories"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAuthService(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	err = db.AutoMigrate(&models.User{}, &models.Note{})
	assert.NoError(t, err)

	cfg := &config.Config{
		JWTSecret:     "testsecret",
		JWTExpiration: 15 * time.Minute,
	}
	userRepo := repositories.NewUserRepository(db)
	authService := NewAuthService(userRepo, cfg)

	t.Run("Register", func(t *testing.T) {
		user, err := authService.Register("testuser", "testpass", "user")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "user", user.Role)
		assert.NotEmpty(t, user.Password)

		dbUser, err := userRepo.FindUserByUsername("testuser")
		assert.NoError(t, err)
		assert.Equal(t, user.ID, dbUser.ID)
	})

	t.Run("Login", func(t *testing.T) {
		_, err := authService.Register("testuser2", "testpass", "user")
		assert.NoError(t, err)

		token, err := authService.Login("testuser2", "testpass")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		_, err = authService.Login("testuser2", "wrongpass")
		assert.Error(t, err)
		assert.Equal(t, "contraseña incorrecta", err.Error())

		_, err = authService.Login("nonexistent", "testpass")
		assert.Error(t, err)
		assert.Equal(t, "usuario no encontrado", err.Error())
	})
}
