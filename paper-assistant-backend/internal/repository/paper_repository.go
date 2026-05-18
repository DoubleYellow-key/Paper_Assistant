package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"paper-assistant-backend/internal/model"
)

var ErrPaperNotFound = errors.New("paper not found")

type PaperRepository struct {
	db *sql.DB
}

func NewPaperRepository(db *sql.DB) *PaperRepository {
	return &PaperRepository{db: db}
}

func (r *PaperRepository) EnsureSchema(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS papers (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  title VARCHAR(255) NOT NULL,
  file_name VARCHAR(255) NOT NULL,
  file_path VARCHAR(512) NOT NULL,
  file_size BIGINT NOT NULL,
  parse_status VARCHAR(32) NOT NULL,
  parse_error TEXT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_papers_user_id_created_at (user_id, created_at, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`
	_, err := r.db.ExecContext(ctx, ddl)
	return err
}

func (r *PaperRepository) CreateTx(ctx context.Context, tx *sql.Tx, paper model.Paper) (model.Paper, error) {
	const query = `
INSERT INTO papers (user_id, title, file_name, file_path, file_size, parse_status, parse_error, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
`
	res, err := tx.ExecContext(
		ctx,
		query,
		paper.UserID,
		paper.Title,
		paper.FileName,
		paper.FilePath,
		paper.FileSize,
		paper.ParseStatus,
		nullString(paper.ParseError),
		paper.CreatedAt,
		paper.UpdatedAt,
	)
	if err != nil {
		return model.Paper{}, fmt.Errorf("insert paper: %w", err)
	}
	insertID, err := res.LastInsertId()
	if err != nil {
		return model.Paper{}, fmt.Errorf("get inserted paper id: %w", err)
	}
	paper.ID = uint64(insertID)
	return paper, nil
}

func (r *PaperRepository) ListByUser(ctx context.Context, userID uint64) ([]model.Paper, error) {
	const query = `
SELECT id, user_id, title, file_name, file_path, file_size, parse_status, COALESCE(parse_error, ''), created_at, updated_at
FROM papers
WHERE user_id = ?
ORDER BY created_at DESC, id DESC
`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list papers by user: %w", err)
	}
	defer rows.Close()

	var papers []model.Paper
	for rows.Next() {
		var paper model.Paper
		if err := rows.Scan(
			&paper.ID,
			&paper.UserID,
			&paper.Title,
			&paper.FileName,
			&paper.FilePath,
			&paper.FileSize,
			&paper.ParseStatus,
			&paper.ParseError,
			&paper.CreatedAt,
			&paper.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan paper row: %w", err)
		}
		papers = append(papers, paper)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate papers: %w", err)
	}
	return papers, nil
}

func (r *PaperRepository) GetByIDAndUserID(ctx context.Context, paperID, userID uint64) (model.Paper, error) {
	const query = `
SELECT id, user_id, title, file_name, file_path, file_size, parse_status, COALESCE(parse_error, ''), created_at, updated_at
FROM papers
WHERE id = ? AND user_id = ?
LIMIT 1
`
	var paper model.Paper
	if err := r.db.QueryRowContext(ctx, query, paperID, userID).Scan(
		&paper.ID,
		&paper.UserID,
		&paper.Title,
		&paper.FileName,
		&paper.FilePath,
		&paper.FileSize,
		&paper.ParseStatus,
		&paper.ParseError,
		&paper.CreatedAt,
		&paper.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Paper{}, ErrPaperNotFound
		}
		return model.Paper{}, fmt.Errorf("get paper by id and user id: %w", err)
	}
	return paper, nil
}

func nullString(v string) any {
	if v == "" {
		return nil
	}
	return v
}
