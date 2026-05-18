package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"paper-assistant-backend/internal/model"
)

var ErrParseJobNotFound = errors.New("parse job not found")

type ParseJobRepository struct {
	db *sql.DB
}

func NewParseJobRepository(db *sql.DB) *ParseJobRepository {
	return &ParseJobRepository{db: db}
}

func (r *ParseJobRepository) EnsureSchema(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS parse_jobs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  paper_id BIGINT UNSIGNED NOT NULL,
  status VARCHAR(32) NOT NULL,
  progress TINYINT UNSIGNED NOT NULL,
  retry_count TINYINT UNSIGNED NOT NULL,
  max_retries TINYINT UNSIGNED NOT NULL,
  started_at DATETIME NULL,
  finished_at DATETIME NULL,
  error_message TEXT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_parse_jobs_paper_id_created_at (paper_id, created_at, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`
	_, err := r.db.ExecContext(ctx, ddl)
	return err
}

func (r *ParseJobRepository) CreateTx(ctx context.Context, tx *sql.Tx, job model.ParseJob) (model.ParseJob, error) {
	const query = `
INSERT INTO parse_jobs (paper_id, status, progress, retry_count, max_retries, started_at, finished_at, error_message, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`
	res, err := tx.ExecContext(
		ctx,
		query,
		job.PaperID,
		job.Status,
		job.Progress,
		job.RetryCount,
		job.MaxRetries,
		job.StartedAt,
		job.FinishedAt,
		nullString(job.ErrorMsg),
		job.CreatedAt,
		job.UpdatedAt,
	)
	if err != nil {
		return model.ParseJob{}, fmt.Errorf("insert parse job: %w", err)
	}
	insertID, err := res.LastInsertId()
	if err != nil {
		return model.ParseJob{}, fmt.Errorf("get inserted parse job id: %w", err)
	}
	job.ID = uint64(insertID)
	return job, nil
}

func (r *ParseJobRepository) GetLatestByPaperIDAndUserID(ctx context.Context, paperID, userID uint64) (model.ParseJob, error) {
	const query = `
SELECT pj.id, pj.paper_id, pj.status, pj.progress, pj.retry_count, pj.max_retries,
       pj.started_at, pj.finished_at, COALESCE(pj.error_message, ''), pj.created_at, pj.updated_at
FROM parse_jobs pj
INNER JOIN papers p ON p.id = pj.paper_id
WHERE pj.paper_id = ? AND p.user_id = ?
ORDER BY pj.created_at DESC, pj.id DESC
LIMIT 1
`
	var job model.ParseJob
	if err := r.db.QueryRowContext(ctx, query, paperID, userID).Scan(
		&job.ID,
		&job.PaperID,
		&job.Status,
		&job.Progress,
		&job.RetryCount,
		&job.MaxRetries,
		&job.StartedAt,
		&job.FinishedAt,
		&job.ErrorMsg,
		&job.CreatedAt,
		&job.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ParseJob{}, ErrParseJobNotFound
		}
		return model.ParseJob{}, fmt.Errorf("get latest parse job by paper and user: %w", err)
	}
	return job, nil
}
