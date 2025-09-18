package models

import (
	"time"

	"gorm.io/gorm"
)

type Session struct {
	gorm.Model
	UserID       uint      `gorm:"not null;index"`
	Token        string    `gorm:"type:text;not null;uniqueIndex"`
	LastActivity time.Time `gorm:"not null;index"`
	ExpiresAt    time.Time `gorm:"not null;index"`
	UserAgent    string    `gorm:"type:text"`
	IP           string    `gorm:"type:varchar(45)"`
	IsActive     bool      `gorm:"not null;default:true;index"`
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s *Session) UpdateLastActivity() {
	s.LastActivity = time.Now()
}
