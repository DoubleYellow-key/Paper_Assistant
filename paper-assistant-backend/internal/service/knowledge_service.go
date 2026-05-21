package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"paper-assistant-backend/internal/agent"
	"paper-assistant-backend/internal/document"
	"paper-assistant-backend/internal/model"
	"paper-assistant-backend/internal/pkg/config"
	"paper-assistant-backend/internal/repository"

	chromem "github.com/philippgille/chromem-go"
)

type KnowledgeQAService struct {
	paperRepo     *repository.PaperRepository
	agentService  agent.Service
	vectorDB      *chromem.DB
	embeddingFunc chromem.EmbeddingFunc
}

func NewKnowledgeQAService(cfg config.Config, paperRepo *repository.PaperRepository, agentService agent.Service) (*KnowledgeQAService, error) {
	if strings.TrimSpace(cfg.Embedding.APIKey) == "" {
		return nil, agent.ErrMissingAPIKey
	}
	vectorDB, err := chromem.NewPersistentDB(cfg.Vector.Path, false)
	if err != nil {
		return nil, fmt.Errorf("open vector db: %w", err)
	}
	embeddingFunc := chromem.NewEmbeddingFuncOpenAICompat(
		cfg.Embedding.BaseURL,
		cfg.Embedding.APIKey,
		cfg.Embedding.Model,
		nil,
	)
	return &KnowledgeQAService{
		paperRepo:     paperRepo,
		agentService:  agentService,
		vectorDB:      vectorDB,
		embeddingFunc: embeddingFunc,
	}, nil
}

func (s *KnowledgeQAService) Ask(ctx context.Context, userID, paperID uint64, question string) (agent.AskResponse, error) {
	if s.agentService == nil {
		return agent.AskResponse{}, fmt.Errorf("agent service unavailable")
	}
	paper, err := s.paperRepo.GetByIDAndUserID(ctx, paperID, userID)
	if err != nil {
		return agent.AskResponse{}, err
	}
	collection, err := s.ensurePaperCollection(ctx, paper)
	if err != nil {
		return agent.AskResponse{}, err
	}

	results, err := collection.Query(ctx, question, 5, nil, nil)
	if err != nil {
		return agent.AskResponse{}, fmt.Errorf("query vector db: %w", err)
	}
	if len(results) == 0 {
		return agent.AskResponse{
			Answer:     "知识库中暂未检索到与该问题直接相关的论文内容。",
			Confidence: "low",
		}, nil
	}

	prompt := buildRAGPrompt(question, results)
	resp, err := s.agentService.Ask(ctx, agent.AskRequest{
		PaperID:      paperID,
		SystemPrompt: buildKnowledgeQASystemPrompt(),
		Query:        prompt,
	})
	if err != nil {
		return agent.AskResponse{}, err
	}
	resp.Citations = buildCitations(results)
	resp.Confidence = classifyConfidence(results[0].Similarity)
	return resp, nil
}

func (s *KnowledgeQAService) ensurePaperCollection(ctx context.Context, paper model.Paper) (*chromem.Collection, error) {
	collectionName := fmt.Sprintf("paper_%d", paper.ID)
	collection, err := s.vectorDB.GetOrCreateCollection(collectionName, map[string]string{
		"paper_id": strconv.FormatUint(paper.ID, 10),
		"title":    paper.Title,
	}, s.embeddingFunc)
	if err != nil {
		return nil, fmt.Errorf("get or create collection: %w", err)
	}
	if collection.Count() > 0 {
		return collection, nil
	}

	text, err := document.ExtractPDFText(toLocalPaperFilePath(paper.FilePath))
	if err != nil {
		return nil, fmt.Errorf("extract pdf text for vector index: %w", err)
	}
	chunks := document.SplitTextIntoChunks(text, 1200)
	if len(chunks) == 0 {
		return nil, fmt.Errorf("no chunks extracted from paper")
	}

	documents := make([]chromem.Document, 0, len(chunks))
	for i, chunk := range chunks {
		documents = append(documents, chromem.Document{
			ID: fmt.Sprintf("%d_%04d", paper.ID, i),
			Metadata: map[string]string{
				"paper_id":    strconv.FormatUint(paper.ID, 10),
				"title":       paper.Title,
				"chunk_index": strconv.Itoa(i + 1),
			},
			Content: chunk,
		})
	}
	if err := collection.AddDocuments(ctx, documents, 4); err != nil {
		return nil, fmt.Errorf("index paper chunks: %w", err)
	}
	return collection, nil
}

func toLocalPaperFilePath(filePath string) string {
	filePath = strings.TrimPrefix(filePath, "/")
	return filepath.Clean(filePath)
}

func buildKnowledgeQASystemPrompt() string {
	return `你是论文知识问答助手。
请严格根据检索到的论文片段回答问题：
1. 优先基于给定片段作答，不要凭空补充论文中不存在的结论。
2. 如果检索片段不足以支持答案，请明确说明“根据当前检索到的论文内容，无法确认”。
3. 回答尽量简洁、准确。`
}

func buildRAGPrompt(question string, results []chromem.Result) string {
	var b strings.Builder
	b.WriteString("以下是从论文知识库检索到的相关片段：\n\n")
	for i, result := range results {
		b.WriteString(fmt.Sprintf("【片段 %d】\n%s\n\n", i+1, strings.TrimSpace(result.Content)))
	}
	b.WriteString("【用户问题】\n")
	b.WriteString(question)
	return b.String()
}

func buildCitations(results []chromem.Result) []string {
	citations := make([]string, 0, len(results))
	for i, result := range results {
		chunkIndex := result.Metadata["chunk_index"]
		if chunkIndex == "" {
			chunkIndex = strconv.Itoa(i + 1)
		}
		content := strings.TrimSpace(result.Content)
		if len([]rune(content)) > 160 {
			content = string([]rune(content)[:160]) + "..."
		}
		citations = append(citations, fmt.Sprintf("片段 %s（相似度 %.2f）: %s", chunkIndex, result.Similarity, content))
	}
	return citations
}

func classifyConfidence(similarity float32) string {
	switch {
	case similarity >= 0.8:
		return "high"
	case similarity >= 0.55:
		return "medium"
	default:
		return "low"
	}
}
