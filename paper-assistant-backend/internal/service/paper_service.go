package service

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"paper-assistant-backend/internal/model"
)

type UploadPaperInput struct {
	UserID   uint64
	Title    string
	FileName string
	FileSize int64
}

type PaperService struct {
	mu           sync.RWMutex
	nextPaperID  uint64
	nextParseJob uint64
	papers       map[uint64]model.Paper
	parseJobs    map[uint64]model.ParseJob // key: paperID
}

func NewPaperService() *PaperService {
	return &PaperService{
		nextPaperID:  1,
		nextParseJob: 1,
		papers:       make(map[uint64]model.Paper),
		parseJobs:    make(map[uint64]model.ParseJob),
	}
}

func (s *PaperService) Upload(in UploadPaperInput) (model.Paper, model.ParseJob) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	paper := model.Paper{
		ID:          s.nextPaperID,
		UserID:      in.UserID,
		Title:       fallbackTitle(in.Title, in.FileName),
		FileName:    in.FileName,
		FilePath:    "/uploads/" + in.FileName,
		FileSize:    in.FileSize,
		ParseStatus: "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.nextPaperID++
	s.papers[paper.ID] = paper

	job := model.ParseJob{
		ID:         s.nextParseJob,
		PaperID:    paper.ID,
		Status:     "queued",
		Progress:   0,
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	s.nextParseJob++
	s.parseJobs[paper.ID] = job
	return paper, job
}

func (s *PaperService) ListByUser(userID uint64) []model.Paper {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.Paper, 0, len(s.papers))
	for _, p := range s.papers {
		if p.UserID == userID {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func (s *PaperService) GetByID(userID, paperID uint64) (model.Paper, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.papers[paperID]
	if !ok || p.UserID != userID {
		return model.Paper{}, fmt.Errorf("paper not found")
	}
	return p, nil
}

func (s *PaperService) GetLatestParseJob(userID, paperID uint64) (model.ParseJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.papers[paperID]
	if !ok || p.UserID != userID {
		return model.ParseJob{}, fmt.Errorf("paper not found")
	}
	job, ok := s.parseJobs[paperID]
	if !ok {
		return model.ParseJob{}, fmt.Errorf("parse job not found")
	}
	return job, nil
}

func fallbackTitle(title, fileName string) string {
	if title != "" {
		return title
	}
	return fileName
}
