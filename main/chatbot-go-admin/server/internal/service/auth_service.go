package service

import (
	"context"
	"errors"
	"time"

	"chatbot-go-admin/server/internal/repository"
	"chatbot-go-admin/server/internal/util"
	"gorm.io/gorm"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrUserDisabled = errors.New("user disabled")

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{userRepo: userRepo, jwtSecret: jwtSecret}
}

func (s *AuthService) Login(ctx context.Context, username string, password string) (token string, userID string, isAdmin bool, err error) {
	u, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", false, ErrInvalidCredentials
		}
		return "", "", false, err
	}
	if u.Status == 0 {
		return "", "", false, ErrUserDisabled
	}
	if !util.CheckPassword(u.PasswordHash, password) {
		return "", "", false, ErrInvalidCredentials
	}

	tok, err := util.SignToken(s.jwtSecret, u.ID, u.IsAdmin == 1, 7*24*time.Hour)
	if err != nil {
		return "", "", false, err
	}
	return tok, u.ID, u.IsAdmin == 1, nil
}

