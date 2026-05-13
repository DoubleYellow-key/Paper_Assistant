package service

import (
	"database/sql"
	"fmt"
	"time"

	"paper-assistant-backend/internal/model"

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
	db *sql.DB
}

func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Register(in RegisterInput) (model.User, error) {
	var existingID uint64
	err := s.db.QueryRow("SELECT id FROM users WHERE email = ? LIMIT 1", in.Email).Scan(&existingID)
	if err == nil {
		return model.User{}, fmt.Errorf("email already exists")
	}
	if err != nil && err != sql.ErrNoRows {
		return model.User{}, fmt.Errorf("query user by email: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, fmt.Errorf("hash password: %w", err)
	}

	res, err := s.db.Exec(
		"INSERT INTO users (username, email, password_hash, role, created_at) VALUES (?, ?, ?, ?, ?)",
		in.Username, in.Email, string(hash), "user", time.Now(),
	)
	if err != nil {
		return model.User{}, fmt.Errorf("insert user: %w", err)
	}
	newID, err := res.LastInsertId()
	if err != nil {
		return model.User{}, fmt.Errorf("get user id: %w", err)
	}

	user := model.User{
		ID:        uint64(newID),
		Username:  in.Username,
		Email:     in.Email,
		Role:      "user",
		CreatedAt: time.Now(),
	}
	return user, nil
}

func (s *AuthService) Login(in LoginInput) (string, model.User, error) {
	var (
		user model.User
		hash string
	)
	err := s.db.QueryRow(
		"SELECT id, username, email, role, created_at, password_hash FROM users WHERE email = ? LIMIT 1",
		in.Email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt, &hash)
	if err == sql.ErrNoRows {
		return "", model.User{}, fmt.Errorf("invalid credentials")
	}
	if err != nil {
		return "", model.User{}, fmt.Errorf("query user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)); err != nil {
		return "", model.User{}, fmt.Errorf("invalid credentials")
	}
	token := fmt.Sprintf("uid-%d", user.ID)
	return token, user, nil
}

func (s *AuthService) GetUser(userID uint64) (model.User, bool) {
	var user model.User
	err := s.db.QueryRow(
		"SELECT id, username, email, role, created_at FROM users WHERE id = ? LIMIT 1",
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return model.User{}, false
	}
	if err != nil {
		return model.User{}, false
	}
	return user, true
}
