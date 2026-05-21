package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"paper-assistant-backend/internal/model"
)

var ErrTranslationNotFound = errors.New("translation not found")

type TranslationRepository struct {
	db *sql.DB
}

func NewTranslationRepository(db *sql.DB) *TranslationRepository {
	return &TranslationRepository{db: db}
}

func (r *TranslationRepository) EnsureSchema(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS paper_translations (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  paper_id BIGINT UNSIGNED NOT NULL,
  target_language VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL,
  content LONGTEXT NULL,
  error_message TEXT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_paper_language (paper_id, target_language),
  KEY idx_translations_paper_id_updated_at (paper_id, updated_at, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`
	_, err := r.db.ExecContext(ctx, ddl)
	return err
}

func (r *TranslationRepository) Upsert(ctx context.Context, translation model.PaperTranslation) (model.PaperTranslation, error) {
	const query = `
INSERT INTO paper_translations (paper_id, target_language, status, content, error_message, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  id = LAST_INSERT_ID(id),
  status = VALUES(status),
  content = VALUES(content),
  error_message = VALUES(error_message),
  updated_at = VALUES(updated_at)
`
	res, err := r.db.ExecContext(
		ctx,
		query,
		translation.PaperID,
		translation.TargetLanguage,
		translation.Status,
		nullString(translation.Content),
		nullString(translation.ErrorMsg),
		translation.CreatedAt,
		translation.UpdatedAt,
	)
	if err != nil {
		return model.PaperTranslation{}, fmt.Errorf("upsert translation: %w", err)
	}
	insertID, err := res.LastInsertId()
	if err != nil {
		return model.PaperTranslation{}, fmt.Errorf("get translation id: %w", err)
	}
	translation.ID = uint64(insertID)
	return r.GetByPaperIDAndLanguage(ctx, translation.PaperID, translation.TargetLanguage)
}

func (r *TranslationRepository) GetByPaperIDAndLanguage(ctx context.Context, paperID uint64, targetLanguage string) (model.PaperTranslation, error) {
	const query = `
SELECT id, paper_id, target_language, status, COALESCE(content, ''), COALESCE(error_message, ''), created_at, updated_at
FROM paper_translations
WHERE paper_id = ? AND target_language = ?
LIMIT 1
`
	var translation model.PaperTranslation
	if err := r.db.QueryRowContext(ctx, query, paperID, targetLanguage).Scan(
		&translation.ID,
		&translation.PaperID,
		&translation.TargetLanguage,
		&translation.Status,
		&translation.Content,
		&translation.ErrorMsg,
		&translation.CreatedAt,
		&translation.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PaperTranslation{}, ErrTranslationNotFound
		}
		return model.PaperTranslation{}, fmt.Errorf("get translation by paper and language: %w", err)
	}
	return translation, nil
}
