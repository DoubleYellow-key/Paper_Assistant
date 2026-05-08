package service

import (
	"fmt"
	"sync"
	"time"

	"paper-assistant-backend/internal/model"
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
	mu        sync.RWMutex
	nextUser  uint64
	usersByID map[uint64]model.User
	usersByEM map[string]uint64
}

func NewAuthService() *AuthService {
	return &AuthService{
		nextUser:  1,
		usersByID: make(map[uint64]model.User),
		usersByEM: make(map[string]uint64),
	}
}

func (s *AuthService) Register(in RegisterInput) (model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.usersByEM[in.Email]; exists {
		return model.User{}, fmt.Errorf("email already exists")
	}
	now := time.Now()
	user := model.User{
		ID:        s.nextUser,
		Username:  in.Username,
		Email:     in.Email,
		Role:      "user",
		CreatedAt: now,
	}
	s.nextUser++
	s.usersByID[user.ID] = user
	s.usersByEM[user.Email] = user.ID
	return user, nil
}

func (s *AuthService) Login(in LoginInput) (string, model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	userID, ok := s.usersByEM[in.Email]
	if !ok {
		return "", model.User{}, fmt.Errorf("invalid credentials")
	}
	user, ok := s.usersByID[userID]
	if !ok {
		return "", model.User{}, fmt.Errorf("invalid credentials")
	}
	token := fmt.Sprintf("uid-%d", user.ID)
	return token, user, nil
}

func (s *AuthService) GetUser(userID uint64) (model.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.usersByID[userID]
	return user, ok
}
