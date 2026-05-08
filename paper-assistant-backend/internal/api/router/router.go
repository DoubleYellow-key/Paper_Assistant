package router

import (
	"context"
	"log"

	"paper-assistant-backend/internal/agent"
	"paper-assistant-backend/internal/api/handler"
	"paper-assistant-backend/internal/api/middleware"
	"paper-assistant-backend/internal/pkg/config"
	"paper-assistant-backend/internal/service"

	"github.com/gin-gonic/gin"
)

func New(cfg config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(middleware.TraceID())

	authService := service.NewAuthService()
	paperService := service.NewPaperService()
	agentService, err := agent.NewEinoService(context.Background(), cfg.LLM)
	if err != nil {
		log.Printf("init eino service failed, ai endpoints may be unavailable: %v", err)
	}

	authHandler := handler.NewAuthHandler(authService)
	paperHandler := handler.NewPaperHandler(paperService, agentService)

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.AuthJWT(), authHandler.Me)
		}

		papers := api.Group("/papers", middleware.AuthJWT())
		{
			papers.POST("/upload", paperHandler.Upload)
			papers.GET("", paperHandler.List)
			papers.GET("/:id", paperHandler.Detail)
			papers.GET("/:id/parse-jobs/latest", paperHandler.LatestParseJob)
			papers.POST("/:id/qa", paperHandler.QA)
			papers.POST("/:id/summary", paperHandler.Summary)
			papers.POST("/:id/term-explain", paperHandler.TermExplain)
			papers.POST("/compare", paperHandler.Compare)
		}
	}
	return r
}
