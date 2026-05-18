package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"paper-assistant-backend/internal/model"
)

var ErrUserNotFound = errors.New("user not found")
var ErrEmailExists = errors.New("email already exists")

type UserRecord struct {
	User         model.User
	PasswordHash string
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) EnsureSchema(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(64) NOT NULL,
  email VARCHAR(191) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  role VARCHAR(32) NOT NULL DEFAULT 'user',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`
	_, err := r.db.ExecContext(ctx, ddl)
	return err
}

func (r *UserRepository) Create(ctx context.Context, username, email, passwordHash string) (model.User, error) {
	const query = `
INSERT INTO users (username, email, password_hash, role)
VALUES (?, ?, ?, 'user')
`
	res, err := r.db.ExecContext(ctx, query, username, email, passwordHash)
	if err != nil {
		if isDuplicateEntry(err) {
			return model.User{}, ErrEmailExists
		}
		return model.User{}, fmt.Errorf("insert user: %w", err)
	}
	userID, err := res.LastInsertId()
	if err != nil {
		return model.User{}, fmt.Errorf("get inserted user id: %w", err)
	}
	user, _, err := r.GetByID(ctx, uint64(userID))
	return user, err
}

func (r *UserRepository) GetByID(ctx context.Context, userID uint64) (model.User, string, error) {
	const query = `
SELECT id, username, email, password_hash, role, created_at
FROM users
WHERE id = ?
LIMIT 1
`
	return r.scanOne(ctx, query, userID)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (model.User, string, error) {
	const query = `
SELECT id, username, email, password_hash, role, created_at
FROM users
WHERE email = ?
LIMIT 1
`
	return r.scanOne(ctx, query, email)
}

func (r *UserRepository) scanOne(ctx context.Context, query string, arg any) (model.User, string, error) {
	var record UserRecord
	row := r.db.QueryRowContext(ctx, query, arg)
	if err := row.Scan(
		&record.User.ID,
		&record.User.Username,
		&record.User.Email,
		&record.PasswordHash,
		&record.User.Role,
		&record.User.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, "", ErrUserNotFound
		}
		return model.User{}, "", fmt.Errorf("query user: %w", err)
	}
	record.User.CreatedAt = record.User.CreatedAt.In(time.Local)
	return record.User, record.PasswordHash, nil
}

func isDuplicateEntry(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate entry") || strings.Contains(msg, "unique constraint")
}
