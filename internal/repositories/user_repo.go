package repositories

import (
	"time"

	"github.com/ramiroschettino/jwt-auth-api/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) IsUsernameTaken(username string) bool {
	var count int64
	r.db.Model(&models.User{}).Where("username = ?", username).Count(&count)
	return count > 0
}

func (r *UserRepository) InvalidateToken(token string, expiresAt time.Time) error {
	return r.db.Create(&models.InvalidToken{
		Token:     token,
		ExpiresAt: expiresAt,
	}).Error
}

func (r *UserRepository) IsTokenInvalid(token string) bool {
	var count int64
	r.db.Model(&models.InvalidToken{}).
		Where("token = ? AND expires_at > ?", token, time.Now()).
		Count(&count)
	return count > 0
}

func (r *UserRepository) CleanupExpiredTokens() error {
	return r.db.Where("expires_at < ?", time.Now()).
		Delete(&models.InvalidToken{}).Error
}

func (r *UserRepository) InvalidateUserTokens(userID uint) error {
	if err := r.CleanupExpiredTokens(); err != nil {
		return err
	}

	return r.db.Model(&models.InvalidToken{}).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Updates(map[string]interface{}{
			"expires_at": time.Now(),
			"reason":     "user_logged_in",
		}).Error
}
