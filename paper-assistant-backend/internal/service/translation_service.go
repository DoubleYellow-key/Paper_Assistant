package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"paper-assistant-backend/internal/agent"
	"paper-assistant-backend/internal/document"
	"paper-assistant-backend/internal/model"
	"paper-assistant-backend/internal/repository"
)

const defaultTargetLanguage = "zh-CN"

type TranslateInput struct {
	UserID          uint64
	PaperID         uint64
	TargetLanguage  string
	ForceRegenerate bool
}

type TranslationService struct {
	paperRepo       *repository.PaperRepository
	translationRepo *repository.TranslationRepository
	agentService    agent.Service
}

func NewTranslationService(
	paperRepo *repository.PaperRepository,
	translationRepo *repository.TranslationRepository,
	agentService agent.Service,
) *TranslationService {
	return &TranslationService{
		paperRepo:       paperRepo,
		translationRepo: translationRepo,
		agentService:    agentService,
	}
}

func (s *TranslationService) Translate(ctx context.Context, in TranslateInput) (model.PaperTranslation, error) {
	if s.agentService == nil {
		return model.PaperTranslation{}, fmt.Errorf("agent service unavailable")
	}
	targetLanguage := normalizeTargetLanguage(in.TargetLanguage)
	if !in.ForceRegenerate {
		translation, err := s.translationRepo.GetByPaperIDAndLanguage(ctx, in.PaperID, targetLanguage)
		if err == nil && translation.Status == "completed" && strings.TrimSpace(translation.Content) != "" {
			return translation, nil
		}
		if err != nil && err != repository.ErrTranslationNotFound {
			return model.PaperTranslation{}, err
		}
	}

	paper, err := s.paperRepo.GetByIDAndUserID(ctx, in.PaperID, in.UserID)
	if err != nil {
		return model.PaperTranslation{}, err
	}

	processing := model.PaperTranslation{
		PaperID:        in.PaperID,
		TargetLanguage: targetLanguage,
		Status:         "processing",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	processing, err = s.translationRepo.Upsert(ctx, processing)
	if err != nil {
		return model.PaperTranslation{}, err
	}

	text, err := document.ExtractPDFText(toLocalFilePath(paper.FilePath))
	if err != nil {
		return s.fail(ctx, processing, fmt.Errorf("extract pdf text: %w", err))
	}
	if strings.TrimSpace(text) == "" {
		return s.fail(ctx, processing, fmt.Errorf("extracted pdf text is empty"))
	}

	chunks := document.SplitTextIntoChunks(text, 2800)
	results := make([]string, 0, len(chunks))
	for index, chunk := range chunks {
		resp, err := s.agentService.Ask(ctx, agent.AskRequest{
			PaperID:      in.PaperID,
			SystemPrompt: buildTranslationSystemPrompt(targetLanguage),
			Query:        fmt.Sprintf("第 %d/%d 段原文如下，请直接输出译文：\n\n%s", index+1, len(chunks), chunk),
		})
		if err != nil {
			return s.fail(ctx, processing, err)
		}
		results = append(results, strings.TrimSpace(resp.Answer))
	}

	processing.Status = "completed"
	processing.Content = strings.TrimSpace(strings.Join(results, "\n\n"))
	processing.ErrorMsg = ""
	processing.UpdatedAt = time.Now()
	return s.translationRepo.Upsert(ctx, processing)
}

func (s *TranslationService) GetLatest(ctx context.Context, userID, paperID uint64, targetLanguage string) (model.PaperTranslation, error) {
	if _, err := s.paperRepo.GetByIDAndUserID(ctx, paperID, userID); err != nil {
		return model.PaperTranslation{}, err
	}
	return s.translationRepo.GetByPaperIDAndLanguage(ctx, paperID, normalizeTargetLanguage(targetLanguage))
}

func (s *TranslationService) fail(ctx context.Context, translation model.PaperTranslation, err error) (model.PaperTranslation, error) {
	translation.Status = "failed"
	translation.ErrorMsg = err.Error()
	translation.Content = ""
	translation.UpdatedAt = time.Now()
	saved, saveErr := s.translationRepo.Upsert(ctx, translation)
	if saveErr != nil {
		return model.PaperTranslation{}, saveErr
	}
	return saved, err
}

func normalizeTargetLanguage(targetLanguage string) string {
	targetLanguage = strings.TrimSpace(targetLanguage)
	if targetLanguage == "" {
		return defaultTargetLanguage
	}
	return targetLanguage
}

func toLocalFilePath(filePath string) string {
	filePath = strings.TrimPrefix(filePath, "/")
	return filepath.Clean(filePath)
}

func buildTranslationSystemPrompt(targetLanguage string) string {
	return fmt.Sprintf(`你是论文翻译助手，请将用户提供的英文论文内容翻译成 %s。
要求：
1. 保留原有标题层级、段落结构、公式、变量名、引用编号。
2. 专业术语首次出现时可保留英文原词。
3. 只输出译文，不要额外解释。`, targetLanguage)
}
