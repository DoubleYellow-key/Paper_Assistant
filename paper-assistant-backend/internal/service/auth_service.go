package service

import (
	"context"
	"fmt"
	"strings"

	"paper-assistant-backend/internal/model"
	"paper-assistant-backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Register(in RegisterInput) (model.User, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, fmt.Errorf("hash password: %w", err)
	}
	user, err := s.userRepo.Create(
		context.Background(),
		strings.TrimSpace(in.Username),
		strings.TrimSpace(strings.ToLower(in.Email)),
		string(passwordHash),
	)
	if err != nil {
		if err == repository.ErrEmailExists {
			return model.User{}, err
		}
		return model.User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (s *AuthService) Login(in LoginInput) (string, model.User, error) {
	user, passwordHash, err := s.userRepo.GetByEmail(context.Background(), strings.TrimSpace(strings.ToLower(in.Email)))
	if err != nil {
		if err == repository.ErrUserNotFound {
			return "", model.User{}, fmt.Errorf("invalid credentials")
		}
		return "", model.User{}, fmt.Errorf("query user by email: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(in.Password)); err != nil {
		return "", model.User{}, fmt.Errorf("invalid credentials")
	}
	token := fmt.Sprintf("uid-%d", user.ID)
	return token, user, nil
}

func (s *AuthService) GetUser(userID uint64) (model.User, bool) {
	user, _, err := s.userRepo.GetByID(context.Background(), userID)
	return user, err == nil
}
