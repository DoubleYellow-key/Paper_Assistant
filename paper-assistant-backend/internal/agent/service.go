package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"paper-assistant-backend/internal/pkg/config"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

var ErrMissingAPIKey = errors.New("missing llm api key")

type AskRequest struct {
	PaperID uint64
	Query   string
}

type AskResponse struct {
	Answer     string   `json:"answer"`
	Citations  []string `json:"citations,omitempty"`
	Confidence string   `json:"confidence,omitempty"`
}

// Service 是 AI 能力抽象层，后续在实现里接入 Eino。
type Service interface {
	Ask(ctx context.Context, req AskRequest) (AskResponse, error)
}

type einoService struct {
	model *openai.ChatModel
}

func NewEinoService(ctx context.Context, cfg config.LLMConfig) (Service, error) {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, ErrMissingAPIKey
	}

	model, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: cfg.BaseURL,
		APIKey:  cfg.APIKey,
		Model:   cfg.Model,
	})
	if err != nil {
		return nil, fmt.Errorf("create eino chat model: %w", err)
	}

	return &einoService{model: model}, nil
}

func (s *einoService) Ask(ctx context.Context, req AskRequest) (AskResponse, error) {
	msg, err := s.model.Generate(ctx, []*schema.Message{
		schema.SystemMessage("你是论文阅读助手，请给出准确、简洁、可追溯的回答。"),
		schema.UserMessage(req.Query),
	})
	if err != nil {
		return AskResponse{}, fmt.Errorf("eino generate: %w", err)
	}
	return AskResponse{
		Answer:     msg.Content,
		Confidence: "medium",
	}, nil
}
