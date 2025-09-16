package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ramiroschettino/jwt-auth-api/internal/config"
	"github.com/ramiroschettino/jwt-auth-api/internal/models"
	"github.com/ramiroschettino/jwt-auth-api/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repositories.UserRepository
	Cfg      *config.Config
}

func NewAuthService(userRepo *repositories.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{userRepo: userRepo, Cfg: cfg}
}

var (
	ErrUserExists       = errors.New("username already exists")
	ErrInvalidUser      = errors.New("invalid username or password")
	ErrTokenInvalid     = errors.New("token is invalid")
	ErrTokenExpired     = errors.New("token has expired")
	ErrTokenBlacklisted = errors.New("token has been invalidated")
)

func (s *AuthService) Register(username, password, role string) (*models.User, error) {
	if s.userRepo.IsUsernameTaken(username) {
		return nil, ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username: username,
		Password: string(hashedPassword),
		Role:     role,
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Login(username, password string) (string, error) {
	user, err := s.userRepo.FindUserByUsername(username)
	if err != nil {
		return "", ErrInvalidUser
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", ErrInvalidUser
	}

	// Invalidar tokens anteriores del usuario
	if err := s.userRepo.InvalidateUserTokens(user.ID); err != nil {
		return "", err
	}

	// Generar nuevo token
	tokenString, err := s.generateToken(user)
	if err != nil {
		return "", err
	}

	// Registrar el nuevo token
	expiresAt := time.Now().Add(s.Cfg.JWTExpiration)
	if err := s.userRepo.InvalidateToken(tokenString, expiresAt); err != nil {
		return "", err
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
		return 0, ErrTokenBlacklisted
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return []byte(s.Cfg.JWTSecret), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Verificar expiración
		if exp, ok := claims["exp"].(float64); ok {
			if float64(time.Now().Unix()) > exp {
				return 0, ErrTokenExpired
			}
		}

		// Convertir user_id a uint
		userID := uint(claims["user_id"].(float64))
		return userID, nil
	}

	return 0, ErrTokenInvalid
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
