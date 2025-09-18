package repositories

import (
	"time"

	"github.com/ramiroschettino/jwt-auth-api/internal/models"
	"gorm.io/gorm"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) CreateSession(session *models.Session) error {
	return r.db.Create(session).Error
}

func (r *SessionRepository) GetActiveSessionByToken(token string) (*models.Session, error) {
	var session models.Session
	err := r.db.Where("token = ? AND is_active = ? AND expires_at > ?", token, true, time.Now()).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) GetActiveSessionsByUserID(userID uint) ([]models.Session, error) {
	var sessions []models.Session
	err := r.db.Where("user_id = ? AND is_active = ? AND expires_at > ?", userID, true, time.Now()).Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *SessionRepository) DeactivateSession(token string) error {
	return r.db.Model(&models.Session{}).
		Where("token = ?", token).
		Updates(map[string]interface{}{
			"is_active":  false,
			"expires_at": time.Now(),
		}).Error
}

func (r *SessionRepository) DeactivateUserSessions(userID uint) error {
	var sessions []models.Session
	if err := r.db.Where("user_id = ? AND is_active = ? AND expires_at > ?", userID, true, time.Now()).Find(&sessions).Error; err != nil {
		return err
	}

	if len(sessions) == 0 {
		return nil
	}

	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&models.Session{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Updates(map[string]interface{}{
			"is_active":  false,
			"expires_at": time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, session := range sessions {
		invalidToken := &models.InvalidToken{
			Token:     session.Token,
			ExpiresAt: session.ExpiresAt,
			UserID:    userID,
			Reason:    "new_login",
		}
		if err := tx.Create(invalidToken).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// Desactiva todas las sesiones activas y agrega sus tokens a la lista negra
func (r *SessionRepository) DeactivateUserSessionsAndBlacklist(userID uint, userRepo *UserRepository) error {
	var sessions []models.Session
	if err := r.db.Where("user_id = ? AND is_active = ? AND expires_at > ?", userID, true, time.Now()).Find(&sessions).Error; err != nil {
		return err
	}

	if len(sessions) == 0 {
		return nil
	}

	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&models.Session{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Updates(map[string]interface{}{
			"is_active":  false,
			"expires_at": time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, session := range sessions {
		if err := userRepo.InvalidateToken(session.Token, session.ExpiresAt); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (r *SessionRepository) UpdateLastActivity(token string) error {
	return r.db.Model(&models.Session{}).
		Where("token = ? AND is_active = ?", token, true).
		Update("last_activity", time.Now()).Error
}

func (r *SessionRepository) CleanupExpiredSessions() error {
	return r.db.Where("expires_at < ?", time.Now()).
		Delete(&models.Session{}).Error
}
