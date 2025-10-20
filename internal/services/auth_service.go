package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ramiroschettino/jwt-auth-api/internal/config"
	apperrors "github.com/ramiroschettino/jwt-auth-api/internal/errors"
	"github.com/ramiroschettino/jwt-auth-api/internal/models"
	"github.com/ramiroschettino/jwt-auth-api/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

// AuthService gestiona la autenticación y sesiones de usuarios
type AuthService struct {
	userRepo    *repositories.UserRepository
	sessionRepo *repositories.SessionRepository
	Cfg         *config.Config
}

// Límite de sesiones simultáneas por usuario
const maxSessionsPerUser = 5

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // segundos hasta expiración
}

func NewAuthService(userRepo *repositories.UserRepository, sessionRepo *repositories.SessionRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		Cfg:         cfg,
	}
}

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

func (s *AuthService) Login(username, password string, userAgent, ip string) (*TokenPair, error) {
	user, err := s.userRepo.FindUserByUsername(username)
	if err != nil {
		return nil, apperrors.ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, apperrors.ErrInvalidPassword
	}

	activeSessions, err := s.sessionRepo.GetActiveSessionsByUserID(user.ID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener sesiones activas: %w", err)
	}
	if len(activeSessions) >= maxSessionsPerUser {
		// Desactivar la sesión más antigua
		oldest := activeSessions[0]
		for _, sess := range activeSessions {
			if sess.CreatedAt.Before(oldest.CreatedAt) {
				oldest = sess
			}
		}
		if err := s.sessionRepo.DeactivateSession(oldest.Token); err != nil {
			return nil, fmt.Errorf("error al desactivar la sesión más antigua: %w", err)
		}
	}

	// Generar par de tokens
	tokenPair, err := s.generateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("error al generar tokens: %w", err)
	}

	// Crear nueva sesión
	session := &models.Session{
		UserID:           user.ID,
		Token:            tokenPair.AccessToken,
		RefreshToken:     tokenPair.RefreshToken,
		LastActivity:     time.Now(),
		ExpiresAt:        time.Now().Add(s.Cfg.JWTExpiration),
		RefreshExpiresAt: time.Now().Add(s.Cfg.RefreshExpiration),
		UserAgent:        userAgent,
		IP:               ip,
		IsActive:         true,
	}

	if err := s.sessionRepo.CreateSession(session); err != nil {
		return nil, fmt.Errorf("error al crear sesión: %w", err)
	}

	return tokenPair, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (*TokenPair, error) {
	// Verificar si el refresh token está en la lista negra
	if s.userRepo.IsTokenInvalid(refreshToken) {
		return nil, apperrors.ErrTokenBlacklisted
	}

	// Buscar la sesión activa con el refresh token
	session, err := s.sessionRepo.GetActiveSessionByRefreshToken(refreshToken)
	if err != nil || session == nil {
		return nil, apperrors.ErrTokenInvalid
	}

	// Verificar si el refresh token está expirado
	if session.IsRefreshExpired() || !session.IsActive {
		return nil, apperrors.ErrTokenExpired
	}

	// Parsear el refresh token para obtener la información del usuario
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return []byte(s.Cfg.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, apperrors.WrapError(err, "refresh token inválido")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, apperrors.ErrTokenInvalid
	}

	// Verificar expiración del refresh token
	if exp, ok := claims["exp"].(float64); ok {
		if float64(time.Now().Unix()) > exp {
			return nil, apperrors.ErrTokenExpired
		}
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, apperrors.WrapError(apperrors.ErrTokenInvalid, "missing user_id claim")
	}

	// Obtener información del usuario
	user, err := s.userRepo.FindUserByID(uint(userID))
	if err != nil {
		return nil, apperrors.ErrUserNotFound
	}

	// Generar nuevo par de tokens
	newTokenPair, err := s.generateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("error al generar nuevos tokens: %w", err)
	}

	// Invalidar el refresh token viejo
	expiresAt := time.Now().Add(s.Cfg.RefreshExpiration)
	if err := s.userRepo.InvalidateToken(refreshToken, expiresAt); err != nil {
		return nil, fmt.Errorf("error al invalidar refresh token antiguo: %w", err)
	}

	// Actualizar la sesión con los nuevos tokens
	newExpiresAt := time.Now().Add(s.Cfg.JWTExpiration)
	newRefreshExpiresAt := time.Now().Add(s.Cfg.RefreshExpiration)

	if err := s.sessionRepo.UpdateTokens(
		refreshToken,
		newTokenPair.AccessToken,
		newTokenPair.RefreshToken,
		newExpiresAt,
		newRefreshExpiresAt,
	); err != nil {
		return nil, fmt.Errorf("error al actualizar sesión: %w", err)
	}

	return newTokenPair, nil
}

func (s *AuthService) Logout(tokenStr string) error {
	// Verificar si la sesión existe y está activa
	session, err := s.sessionRepo.GetActiveSessionByToken(tokenStr)
	if err != nil || session == nil {
		// Si no existe la sesión, igual intentamos invalidar el token
		expiresAt := time.Now().Add(s.Cfg.JWTExpiration)
		return s.userRepo.InvalidateToken(tokenStr, expiresAt)
	}

	// Desactivar la sesión
	if err := s.sessionRepo.DeactivateSession(tokenStr); err != nil {
		return fmt.Errorf("error al desactivar la sesión: %w", err)
	}

	// Invalidar ambos tokens (access y refresh)
	accessExpiresAt := time.Now().Add(s.Cfg.JWTExpiration)
	if err := s.userRepo.InvalidateToken(tokenStr, accessExpiresAt); err != nil {
		return fmt.Errorf("error al invalidar access token: %w", err)
	}

	refreshExpiresAt := time.Now().Add(s.Cfg.RefreshExpiration)
	if err := s.userRepo.InvalidateToken(session.RefreshToken, refreshExpiresAt); err != nil {
		return fmt.Errorf("error al invalidar refresh token: %w", err)
	}

	return nil
}

func (s *AuthService) ValidateToken(tokenStr string) (uint, string, error) {
	// Verificar si el token está en la lista negra
	if s.userRepo.IsTokenInvalid(tokenStr) {
		return 0, "", apperrors.ErrTokenBlacklisted
	}

	// Verificar si la sesión está activa
	session, err := s.sessionRepo.GetActiveSessionByToken(tokenStr)
	if err != nil || session == nil || session.IsExpired() || !session.IsActive {
		return 0, "", apperrors.ErrTokenInvalid
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return []byte(s.Cfg.JWTSecret), nil
	})

	if err != nil {
		return 0, "", apperrors.WrapError(err, "failed to parse token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if exp, ok := claims["exp"].(float64); ok {
			if float64(time.Now().Unix()) > exp {
				return 0, "", apperrors.ErrTokenExpired
			}
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			return 0, "", apperrors.WrapError(apperrors.ErrTokenInvalid, "missing user_id claim")
		}

		role, ok := claims["role"].(string)
		if !ok {
			return 0, "", apperrors.WrapError(apperrors.ErrTokenInvalid, "missing role claim")
		}

		// Actualizar la última actividad de la sesión
		_ = s.sessionRepo.UpdateLastActivity(tokenStr)

		return uint(userID), role, nil
	}

	return 0, "", apperrors.ErrTokenInvalid
}

func (s *AuthService) generateTokenPair(user *models.User) (*TokenPair, error) {
	// Generar Access Token
	accessToken, err := s.generateToken(user, s.Cfg.JWTExpiration, "access")
	if err != nil {
		return nil, fmt.Errorf("error al generar access token: %w", err)
	}

	// Generar Refresh Token
	refreshToken, err := s.generateToken(user, s.Cfg.RefreshExpiration, "refresh")
	if err != nil {
		return nil, fmt.Errorf("error al generar refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.Cfg.JWTExpiration.Seconds()),
	}, nil
}

func (s *AuthService) generateToken(user *models.User, expiration time.Duration, tokenType string) (string, error) {
	jti := make([]byte, 16)
	_, err := rand.Read(jti)
	if err != nil {
		return "", err
	}
	jtiStr := hex.EncodeToString(jti)

	claims := jwt.MapClaims{
		"user_id":    user.ID,
		"username":   user.Username,
		"role":       user.Role,
		"exp":        time.Now().Add(expiration).Unix(),
		"iat":        time.Now().Unix(),
		"jti":        jtiStr,
		"token_type": tokenType,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.Cfg.JWTSecret))
}
