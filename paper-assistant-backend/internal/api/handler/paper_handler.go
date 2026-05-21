package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"paper-assistant-backend/internal/agent"
	"paper-assistant-backend/internal/api/middleware"
	apperrors "paper-assistant-backend/internal/pkg/errors"
	"paper-assistant-backend/internal/pkg/response"
	"paper-assistant-backend/internal/repository"
	"paper-assistant-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type PaperHandler struct {
	knowledgeQAService *service.KnowledgeQAService
	paperService       *service.PaperService
	translationService *service.TranslationService
	agentService       agent.Service
}

func NewPaperHandler(
	knowledgeQAService *service.KnowledgeQAService,
	paperService *service.PaperService,
	translationService *service.TranslationService,
	agentService agent.Service,
) *PaperHandler {
	return &PaperHandler{
		knowledgeQAService: knowledgeQAService,
		paperService:       paperService,
		translationService: translationService,
		agentService:       agentService,
	}
}

func (h *PaperHandler) Upload(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "unauthorized")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, apperrors.CodeBadRequest, "file is required")
		return
	}
	if err := os.MkdirAll("uploads", 0755); err != nil {
		response.Fail(c, http.StatusInternalServerError, apperrors.CodeInternal, "create upload dir failed")
		return
	}
	storedName := buildStoredFileName(file.Filename)
	dst := filepath.Join("uploads", storedName)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.Fail(c, http.StatusInternalServerError, apperrors.CodeInternal, "save uploaded file failed")
		return
	}
	title := c.PostForm("title")
	paper, job, err := h.paperService.Upload(service.UploadPaperInput{
		UserID:   userID,
		Title:    title,
		FileName: file.Filename,
		FilePath: "/uploads/" + storedName,
		FileSize: file.Size,
	})
	if err != nil {
		_ = os.Remove(dst)
		response.Fail(c, http.StatusInternalServerError, apperrors.CodeInternal, "create paper record failed")
		return
	}
	response.OK(c, gin.H{
		"paper":     paper,
		"parse_job": job,
	})
}

func (h *PaperHandler) List(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "unauthorized")
		return
	}
	items, err := h.paperService.ListByUser(userID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, apperrors.CodeInternal, "list papers failed")
		return
	}
	response.OK(c, gin.H{"items": items})
}

func (h *PaperHandler) Detail(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "unauthorized")
		return
	}
	paperID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || paperID == 0 {
		response.Fail(c, http.StatusBadRequest, apperrors.CodeBadRequest, "invalid id")
		return
	}
	paper, err := h.paperService.GetByID(userID, paperID)
	if err != nil {
		response.Fail(c, http.StatusNotFound, apperrors.CodeNotFound, "paper not found")
		return
	}
	response.OK(c, gin.H{"paper": paper})
}

func (h *PaperHandler) LatestParseJob(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "unauthorized")
		return
	}
	paperID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || paperID == 0 {
		response.Fail(c, http.StatusBadRequest, apperrors.CodeBadRequest, "invalid id")
		return
	}
	job, err := h.paperService.GetLatestParseJob(userID, paperID)
	if err != nil {
		response.Fail(c, http.StatusNotFound, apperrors.CodeNotFound, "parse job not found")
		return
	}
	response.OK(c, gin.H{"parse_job": job})
}

func (h *PaperHandler) QA(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "unauthorized")
		return
	}
	if h.knowledgeQAService == nil {
		response.Fail(c, http.StatusServiceUnavailable, apperrors.CodeModelFailed, "knowledge qa service unavailable")
		return
	}
	paperID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || paperID == 0 {
		response.Fail(c, http.StatusBadRequest, apperrors.CodeBadRequest, "invalid id")
		return
	}
	var req askRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apperrors.CodeBadRequest, "invalid request body")
		return
	}
	resp, err := h.knowledgeQAService.Ask(c.Request.Context(), userID, paperID, req.Query)
	if err != nil {
		if errors.Is(err, agent.ErrMissingAPIKey) {
			response.Fail(c, http.StatusServiceUnavailable, apperrors.CodeModelFailed, "missing llm api key")
			return
		}
		if errors.Is(err, repository.ErrPaperNotFound) {
			response.Fail(c, http.StatusNotFound, apperrors.CodeNotFound, "paper not found")
			return
		}
		response.Fail(c, http.StatusBadGateway, apperrors.CodeModelFailed, err.Error())
		return
	}
	response.OK(c, resp)
}

func (h *PaperHandler) Summary(c *gin.Context) {
	h.askByPrompt(c, "请生成该论文的结构化摘要：")
}

func (h *PaperHandler) TermExplain(c *gin.Context) {
	h.askByPrompt(c, "请解释术语并给出通俗说明：")
}

func (h *PaperHandler) Translate(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "unauthorized")
		return
	}
	if h.translationService == nil {
		response.Fail(c, http.StatusServiceUnavailable, apperrors.CodeModelFailed, "translation service unavailable")
		return
	}
	paperID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || paperID == 0 {
		response.Fail(c, http.StatusBadRequest, apperrors.CodeBadRequest, "invalid id")
		return
	}
	var req translateRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		response.Fail(c, http.StatusBadRequest, apperrors.CodeBadRequest, "invalid request body")
		return
	}
	translation, err := h.translationService.Translate(c.Request.Context(), service.TranslateInput{
		UserID:          userID,
		PaperID:         paperID,
		TargetLanguage:  req.TargetLanguage,
		ForceRegenerate: req.ForceRegenerate,
	})
	if err != nil {
		if errors.Is(err, agent.ErrMissingAPIKey) {
			response.Fail(c, http.StatusServiceUnavailable, apperrors.CodeModelFailed, "missing llm api key")
			return
		}
		response.Fail(c, http.StatusBadGateway, apperrors.CodeModelFailed, err.Error())
		return
	}
	response.OK(c, gin.H{"translation": translation})
}

func (h *PaperHandler) LatestTranslation(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "unauthorized")
		return
	}
	if h.translationService == nil {
		response.Fail(c, http.StatusServiceUnavailable, apperrors.CodeModelFailed, "translation service unavailable")
		return
	}
	paperID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || paperID == 0 {
		response.Fail(c, http.StatusBadRequest, apperrors.CodeBadRequest, "invalid id")
		return
	}
	translation, err := h.translationService.GetLatest(c.Request.Context(), userID, paperID, c.Query("target_language"))
	if err != nil {
		if errors.Is(err, repository.ErrTranslationNotFound) {
			response.OK(c, gin.H{"translation": nil})
			return
		}
		if errors.Is(err, repository.ErrPaperNotFound) {
			response.Fail(c, http.StatusNotFound, apperrors.CodeNotFound, "paper not found")
			return
		}
		response.Fail(c, http.StatusInternalServerError, apperrors.CodeInternal, "get translation failed")
		return
	}
	response.OK(c, gin.H{"translation": translation})
}

func (h *PaperHandler) Compare(c *gin.Context) {
	response.OK(c, gin.H{"message": "TODO: compare endpoint"})
}

type askRequest struct {
	Query string `json:"query" binding:"required"`
}

type translateRequest struct {
	TargetLanguage  string `json:"target_language"`
	ForceRegenerate bool   `json:"force_regenerate"`
}

func (h *PaperHandler) askByPrompt(c *gin.Context, instruction string) {
	if h.agentService == nil {
		response.Fail(c, http.StatusServiceUnavailable, apperrors.CodeModelFailed, "agent service unavailable")
		return
	}

	paperID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || paperID == 0 {
		response.Fail(c, http.StatusBadRequest, apperrors.CodeBadRequest, "invalid id")
		return
	}

	var req askRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apperrors.CodeBadRequest, "invalid request body")
		return
	}

	resp, err := h.agentService.Ask(c.Request.Context(), agent.AskRequest{
		PaperID: paperID,
		Query:   instruction + "\n" + req.Query,
	})
	if err != nil {
		if errors.Is(err, agent.ErrMissingAPIKey) {
			response.Fail(c, http.StatusServiceUnavailable, apperrors.CodeModelFailed, "missing llm api key")
			return
		}
		response.Fail(c, http.StatusBadGateway, apperrors.CodeModelFailed, err.Error())
		return
	}
	response.OK(c, resp)
}

func buildStoredFileName(original string) string {
	base := filepath.Base(original)
	base = strings.ReplaceAll(base, string(os.PathSeparator), "_")
	base = strings.ReplaceAll(base, " ", "_")
	randHex := randomHex(12)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	if name == "" {
		name = "paper"
	}
	if ext == "" {
		ext = ".pdf"
	}
	return fmt.Sprintf("%s_%s%s", name, randHex, ext)
}

func randomHex(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
