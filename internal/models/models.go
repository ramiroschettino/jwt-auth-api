package models

import (
	"time"

	"gorm.io/gorm"
)

// User almacena la información básica de un usuario del sistema
type User struct {
	gorm.Model
	Username string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
	Role     string `gorm:"not null"`
}

// Note representa una nota personal del usuario
type Note struct {
	gorm.Model
	Title   string `gorm:"not null"`
	Content string
	UserID  uint `gorm:"not null"`
}

type TokenClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type InvalidToken struct {
	gorm.Model
	Token     string    `gorm:"type:text;not null;uniqueIndex"`
	ExpiresAt time.Time `gorm:"not null;index"`
	UserID    uint      `gorm:"not null;index"`
	Reason    string    `gorm:"not null"`
}
