package models

import (
	"time"

	"gorm.io/gorm"
)

type InvalidToken struct {
	gorm.Model
	Token     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	UserID    uint      `gorm:"not null"`
	Reason    string
}
