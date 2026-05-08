package handler

import (
	"errors"
	"net/http"
	"strconv"

	"paper-assistant-backend/internal/agent"
	"paper-assistant-backend/internal/api/middleware"
	apperrors "paper-assistant-backend/internal/pkg/errors"
	"paper-assistant-backend/internal/pkg/response"
	"paper-assistant-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type PaperHandler struct {
	paperService *service.PaperService
	agentService agent.Service
}

func NewPaperHandler(paperService *service.PaperService, agentService agent.Service) *PaperHandler {
	return &PaperHandler{
		paperService: paperService,
		agentService: agentService,
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
	title := c.PostForm("title")
	paper, job := h.paperService.Upload(service.UploadPaperInput{
		UserID:   userID,
		Title:    title,
		FileName: file.Filename,
		FileSize: file.Size,
	})
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
	response.OK(c, gin.H{"items": h.paperService.ListByUser(userID)})
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
	h.askByPrompt(c, "请根据论文内容回答用户问题：")
}

func (h *PaperHandler) Summary(c *gin.Context) {
	h.askByPrompt(c, "请生成该论文的结构化摘要：")
}

func (h *PaperHandler) TermExplain(c *gin.Context) {
	h.askByPrompt(c, "请解释术语并给出通俗说明：")
}

func (h *PaperHandler) Compare(c *gin.Context) {
	response.OK(c, gin.H{"message": "TODO: compare endpoint"})
}

type askRequest struct {
	Query string `json:"query" binding:"required"`
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
