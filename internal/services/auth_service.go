package services

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ramiroschettino/jwt-auth-api/internal/config"
	apperrors "github.com/ramiroschettino/jwt-auth-api/internal/errors"
	"github.com/ramiroschettino/jwt-auth-api/internal/models"
	"github.com/ramiroschettino/jwt-auth-api/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo    *repositories.UserRepository
	sessionRepo *repositories.SessionRepository
	Cfg         *config.Config
}

const maxSessionsPerUser = 5

func NewAuthService(userRepo *repositories.UserRepository, sessionRepo *repositories.SessionRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		Cfg:         cfg,
	}
}

// Error variables are now centralized in the errors package

func (s *AuthService) Register(username, password, role string) (*models.User, error) {
	if s.userRepo.IsUsernameTaken(username) {
		return nil, apperrors.ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.WrapError(err, "failed to hash password")
	}

	user := &models.User{
		Username: username,
		Password: string(hashedPassword),
		Role:     role,
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, apperrors.WrapError(err, "failed to create user")
	}
	return user, nil
}

func (s *AuthService) Login(username, password string, userAgent, ip string) (string, error) {
	user, err := s.userRepo.FindUserByUsername(username)
	if err != nil {
		return "", apperrors.ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", apperrors.ErrInvalidPassword
	}

	// Invalida todos los tokens anteriores y desactiva sesiones previas
	if err := s.sessionRepo.DeactivateUserSessionsAndBlacklist(user.ID, s.userRepo); err != nil {
		return "", fmt.Errorf("error al invalidar sesiones anteriores: %w", err)
	}

	// Generar nuevo token
	tokenString, err := s.generateToken(user)
	if err != nil {
		return "", fmt.Errorf("error al generar token: %w", err)
	}

	// Crear nueva sesión
	session := &models.Session{
		UserID:       user.ID,
		Token:        tokenString,
		LastActivity: time.Now(),
		ExpiresAt:    time.Now().Add(s.Cfg.JWTExpiration),
		UserAgent:    userAgent,
		IP:           ip,
		IsActive:     true,
	}

	if err := s.sessionRepo.CreateSession(session); err != nil {
		return "", fmt.Errorf("error al crear sesión: %w", err)
	}

	return tokenString, nil
}

func (s *AuthService) Logout(tokenStr string) error {
	// Invalidar el token actual
	expiresAt := time.Now().Add(s.Cfg.JWTExpiration)
	return s.userRepo.InvalidateToken(tokenStr, expiresAt)
}

func (s *AuthService) ValidateToken(tokenStr string) (uint, error) {
	// Verificar si el token está en la lista negra
	if s.userRepo.IsTokenInvalid(tokenStr) {
		return 0, apperrors.ErrTokenBlacklisted
	}

	// Verificar si la sesión está activa
	session, err := s.sessionRepo.GetActiveSessionByToken(tokenStr)
	if err != nil || session == nil || session.IsExpired() || !session.IsActive {
		return 0, apperrors.ErrTokenInvalid
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return []byte(s.Cfg.JWTSecret), nil
	})

	if err != nil {
		return 0, apperrors.WrapError(err, "failed to parse token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if exp, ok := claims["exp"].(float64); ok {
			if float64(time.Now().Unix()) > exp {
				return 0, apperrors.ErrTokenExpired
			}
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			return 0, apperrors.WrapError(apperrors.ErrTokenInvalid, "missing user_id claim")
		}

		// Actualizar la última actividad de la sesión
		_ = s.sessionRepo.UpdateLastActivity(tokenStr)

		return uint(userID), nil
	}

	return 0, apperrors.ErrTokenInvalid
}

func (s *AuthService) generateToken(user *models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(s.Cfg.JWTExpiration).Unix(),
		"iat":      time.Now().Unix(),
	})

	return token.SignedString([]byte(s.Cfg.JWTSecret))
}
