package service

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"paper-assistant-backend/internal/model"
	"paper-assistant-backend/internal/rag/parser"
)

type UploadPaperInput struct {
	UserID      uint64
	Title       string
	FileName    string
	FilePath    string
	StoragePath string
	FileSize    int64
}

type PaperService struct {
	db *sql.DB
}

func NewPaperService(db *sql.DB) *PaperService {
	return &PaperService{db: db}
}

func (s *PaperService) Upload(in UploadPaperInput) (model.Paper, model.ParseJob) {
	now := time.Now()

	res, err := s.db.Exec(
		`INSERT INTO papers
		(user_id, title, file_name, file_path, storage_path, file_size, parse_status, parse_error, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 'pending', '', ?, ?)`,
		in.UserID,
		fallbackTitle(in.Title, in.FileName),
		in.FileName,
		fallbackFilePath(in.FilePath, in.FileName),
		in.StoragePath,
		in.FileSize,
		now, now,
	)
	if err != nil {
		return model.Paper{}, model.ParseJob{}
	}
	paperID, err := res.LastInsertId()
	if err != nil {
		return model.Paper{}, model.ParseJob{}
	}

	jobRes, err := s.db.Exec(
		`INSERT INTO parse_jobs
		(paper_id, status, progress, retry_count, max_retries, created_at, updated_at)
		VALUES (?, 'queued', 0, 0, 3, ?, ?)`,
		paperID, now, now,
	)
	if err != nil {
		return model.Paper{}, model.ParseJob{}
	}
	jobID, err := jobRes.LastInsertId()
	if err != nil {
		return model.Paper{}, model.ParseJob{}
	}

	paper := model.Paper{
		ID:          uint64(paperID),
		UserID:      in.UserID,
		Title:       fallbackTitle(in.Title, in.FileName),
		FileName:    in.FileName,
		FilePath:    fallbackFilePath(in.FilePath, in.FileName),
		StoragePath: in.StoragePath,
		FileSize:    in.FileSize,
		ParseStatus: "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	job := model.ParseJob{
		ID:         uint64(jobID),
		PaperID:    uint64(paperID),
		Status:     "queued",
		Progress:   0,
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	go s.runParse(paper.ID)
	return paper, job
}

func (s *PaperService) ListByUser(userID uint64) []model.Paper {
	rows, err := s.db.Query(
		`SELECT id, user_id, title, file_name, file_path, storage_path, file_size, parse_status, parse_error, created_at, updated_at
		 FROM papers WHERE user_id = ? ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return []model.Paper{}
	}
	defer rows.Close()

	items := make([]model.Paper, 0)
	for rows.Next() {
		var p model.Paper
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.Title, &p.FileName, &p.FilePath, &p.StoragePath,
			&p.FileSize, &p.ParseStatus, &p.ParseError, &p.CreatedAt, &p.UpdatedAt,
		); err == nil {
			items = append(items, p)
		}
	}
	return items
}

func (s *PaperService) GetByID(userID, paperID uint64) (model.Paper, error) {
	var p model.Paper
	err := s.db.QueryRow(
		`SELECT id, user_id, title, file_name, file_path, storage_path, file_size, parse_status, parse_error, created_at, updated_at
		 FROM papers WHERE id = ? AND user_id = ? LIMIT 1`,
		paperID, userID,
	).Scan(
		&p.ID, &p.UserID, &p.Title, &p.FileName, &p.FilePath, &p.StoragePath,
		&p.FileSize, &p.ParseStatus, &p.ParseError, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return model.Paper{}, fmt.Errorf("paper not found")
	}
	if err != nil {
		return model.Paper{}, fmt.Errorf("query paper: %w", err)
	}
	return p, nil
}

func (s *PaperService) GetLatestParseJob(userID, paperID uint64) (model.ParseJob, error) {
	if _, err := s.GetByID(userID, paperID); err != nil {
		return model.ParseJob{}, fmt.Errorf("paper not found")
	}

	var (
		j          model.ParseJob
		startedAt  sql.NullTime
		finishedAt sql.NullTime
		errorMsg   sql.NullString
	)
	err := s.db.QueryRow(
		`SELECT id, paper_id, status, progress, retry_count, max_retries, started_at, finished_at, error_message, created_at, updated_at
		 FROM parse_jobs WHERE paper_id = ? LIMIT 1`,
		paperID,
	).Scan(
		&j.ID, &j.PaperID, &j.Status, &j.Progress, &j.RetryCount, &j.MaxRetries,
		&startedAt, &finishedAt, &errorMsg, &j.CreatedAt, &j.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return model.ParseJob{}, fmt.Errorf("parse job not found")
	}
	if err != nil {
		return model.ParseJob{}, fmt.Errorf("query parse job: %w", err)
	}
	if startedAt.Valid {
		j.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		j.FinishedAt = &finishedAt.Time
	}
	if errorMsg.Valid {
		j.ErrorMsg = errorMsg.String
	}
	return j, nil
}

func (s *PaperService) GetParsedText(userID, paperID uint64) (string, error) {
	if _, err := s.GetByID(userID, paperID); err != nil {
		return "", fmt.Errorf("paper not found")
	}
	var text string
	err := s.db.QueryRow("SELECT content FROM paper_parsed_texts WHERE paper_id = ? LIMIT 1", paperID).Scan(&text)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("paper not parsed")
	}
	if err != nil {
		return "", fmt.Errorf("query parsed text: %w", err)
	}
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("paper not parsed")
	}
	return text, nil
}

func fallbackTitle(title, fileName string) string {
	if title != "" {
		return title
	}
	return fileName
}

func fallbackFilePath(filePath, fileName string) string {
	if filePath != "" {
		return filePath
	}
	return "/api/v1/uploads/" + fileName
}

func (s *PaperService) runParse(paperID uint64) {
	s.markParseRunning(paperID)

	paper, err := s.getPaperByID(paperID)
	if err != nil {
		s.markParseFailed(paperID, "paper not found")
		return
	}
	if strings.TrimSpace(paper.StoragePath) == "" {
		s.markParseFailed(paperID, "storage path empty")
		return
	}

	text, err := parser.ParseFileText(paper.StoragePath)
	if err != nil {
		s.markParseFailed(paperID, err.Error())
		return
	}
	if strings.TrimSpace(text) == "" {
		s.markParseFailed(paperID, "parsed content is empty")
		return
	}

	now := time.Now()
	_, _ = s.db.Exec(
		`INSERT INTO paper_parsed_texts (paper_id, content, updated_at)
		 VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE content = VALUES(content), updated_at = VALUES(updated_at)`,
		paperID, text, now,
	)
	_, _ = s.db.Exec(
		"UPDATE parse_jobs SET status='success', progress=100, finished_at=?, error_message='', updated_at=? WHERE paper_id=?",
		now, now, paperID,
	)
	_, _ = s.db.Exec(
		"UPDATE papers SET parse_status='done', parse_error='', updated_at=? WHERE id=?",
		now, paperID,
	)
}

func (s *PaperService) markParseRunning(paperID uint64) {
	now := time.Now()
	_, _ = s.db.Exec(
		"UPDATE parse_jobs SET status='running', progress=20, started_at=?, updated_at=? WHERE paper_id=?",
		now, now, paperID,
	)
	_, _ = s.db.Exec(
		"UPDATE papers SET parse_status='processing', updated_at=? WHERE id=?",
		now, paperID,
	)
}

func (s *PaperService) markParseFailed(paperID uint64, reason string) {
	now := time.Now()
	_, _ = s.db.Exec(
		"UPDATE parse_jobs SET status='failed', progress=100, finished_at=?, error_message=?, updated_at=? WHERE paper_id=?",
		now, reason, now, paperID,
	)
	_, _ = s.db.Exec(
		"UPDATE papers SET parse_status='failed', parse_error=?, updated_at=? WHERE id=?",
		reason, now, paperID,
	)
}

func (s *PaperService) getPaperByID(paperID uint64) (model.Paper, error) {
	var p model.Paper
	err := s.db.QueryRow(
		`SELECT id, user_id, title, file_name, file_path, storage_path, file_size, parse_status, parse_error, created_at, updated_at
		 FROM papers WHERE id = ? LIMIT 1`,
		paperID,
	).Scan(
		&p.ID, &p.UserID, &p.Title, &p.FileName, &p.FilePath, &p.StoragePath,
		&p.FileSize, &p.ParseStatus, &p.ParseError, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return model.Paper{}, err
	}
	return p, nil
}
