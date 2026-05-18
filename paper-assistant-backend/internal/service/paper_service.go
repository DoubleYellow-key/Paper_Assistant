package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"paper-assistant-backend/internal/model"
	"paper-assistant-backend/internal/repository"
)

type UploadPaperInput struct {
	UserID   uint64
	Title    string
	FileName string
	FilePath string
	FileSize int64
}

type PaperService struct {
	db           *sql.DB
	paperRepo    *repository.PaperRepository
	parseJobRepo *repository.ParseJobRepository
}

func NewPaperService(db *sql.DB, paperRepo *repository.PaperRepository, parseJobRepo *repository.ParseJobRepository) *PaperService {
	return &PaperService{
		db:           db,
		paperRepo:    paperRepo,
		parseJobRepo: parseJobRepo,
	}
}

func (s *PaperService) Upload(in UploadPaperInput) (model.Paper, model.ParseJob, error) {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Paper{}, model.ParseJob{}, fmt.Errorf("begin upload transaction: %w", err)
	}
	now := time.Now()
	startedAt := now
	finishedAt := now
	paper := model.Paper{
		UserID:      in.UserID,
		Title:       fallbackTitle(in.Title, in.FileName),
		FileName:    in.FileName,
		FilePath:    in.FilePath,
		FileSize:    in.FileSize,
		ParseStatus: "completed",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	paper, err = s.paperRepo.CreateTx(ctx, tx, paper)
	if err != nil {
		_ = tx.Rollback()
		return model.Paper{}, model.ParseJob{}, err
	}

	job := model.ParseJob{
		PaperID:    paper.ID,
		Status:     "completed",
		Progress:   100,
		RetryCount: 0,
		MaxRetries: 3,
		StartedAt:  &startedAt,
		FinishedAt: &finishedAt,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	job, err = s.parseJobRepo.CreateTx(ctx, tx, job)
	if err != nil {
		_ = tx.Rollback()
		return model.Paper{}, model.ParseJob{}, err
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return model.Paper{}, model.ParseJob{}, fmt.Errorf("commit upload transaction: %w", err)
	}
	return paper, job, nil
}

func (s *PaperService) ListByUser(userID uint64) ([]model.Paper, error) {
	papers, err := s.paperRepo.ListByUser(context.Background(), userID)
	if err != nil {
		return nil, err
	}
	return papers, nil
}

func (s *PaperService) GetByID(userID, paperID uint64) (model.Paper, error) {
	paper, err := s.paperRepo.GetByIDAndUserID(context.Background(), paperID, userID)
	if err != nil {
		return model.Paper{}, err
	}
	return paper, nil
}

func (s *PaperService) GetLatestParseJob(userID, paperID uint64) (model.ParseJob, error) {
	job, err := s.parseJobRepo.GetLatestByPaperIDAndUserID(context.Background(), paperID, userID)
	if err != nil {
		return model.ParseJob{}, err
	}
	return job, nil
}

func fallbackTitle(title, fileName string) string {
	title = strings.TrimSpace(title)
	if title != "" {
		return title
	}
	return fileName
}
