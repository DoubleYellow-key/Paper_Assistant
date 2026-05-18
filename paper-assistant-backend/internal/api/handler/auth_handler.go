package handler

import (
	"errors"
	"net/http"

	"paper-assistant-backend/internal/api/middleware"
	apperrors "paper-assistant-backend/internal/pkg/errors"
	"paper-assistant-backend/internal/pkg/response"
	"paper-assistant-backend/internal/repository"
	"paper-assistant-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apperrors.CodeBadRequest, "invalid request body")
		return
	}
	user, err := h.authService.Register(req)
	if err != nil {
		if errors.Is(err, repository.ErrEmailExists) {
			response.Fail(c, http.StatusConflict, apperrors.CodeStateConflict, err.Error())
			return
		}
		response.Fail(c, http.StatusInternalServerError, apperrors.CodeInternal, "register failed")
		return
	}
	response.OK(c, gin.H{"user": user})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apperrors.CodeBadRequest, "invalid request body")
		return
	}
	token, user, err := h.authService.Login(req)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "invalid credentials")
		return
	}
	response.OK(c, gin.H{"token": token, "user": user})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "unauthorized")
		return
	}
	user, ok := h.authService.GetUser(userID)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "unauthorized")
		return
	}
	response.OK(c, gin.H{"user": user})
}
