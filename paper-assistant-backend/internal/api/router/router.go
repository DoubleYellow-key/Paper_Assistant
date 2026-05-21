package router

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"paper-assistant-backend/internal/agent"
	"paper-assistant-backend/internal/api/handler"
	"paper-assistant-backend/internal/api/middleware"
	"paper-assistant-backend/internal/pkg/config"
	"paper-assistant-backend/internal/repository"
	"paper-assistant-backend/internal/service"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func New(cfg config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(middleware.TraceID())
	r.StaticFS("/api/v1/uploads", http.Dir("./uploads"))

	if cfg.MySQL.DSN == "" {
		log.Fatal("MYSQL_DSN is required")
	}
	db, err := sql.Open("mysql", cfg.MySQL.DSN)
	if err != nil {
		log.Fatalf("open mysql: %v", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		log.Fatalf("ping mysql: %v", err)
	}
	userRepo := repository.NewUserRepository(db)
	if err := userRepo.EnsureSchema(context.Background()); err != nil {
		log.Fatalf("ensure user schema: %v", err)
	}
	paperRepo := repository.NewPaperRepository(db)
	if err := paperRepo.EnsureSchema(context.Background()); err != nil {
		log.Fatalf("ensure paper schema: %v", err)
	}
	parseJobRepo := repository.NewParseJobRepository(db)
	if err := parseJobRepo.EnsureSchema(context.Background()); err != nil {
		log.Fatalf("ensure parse job schema: %v", err)
	}
	translationRepo := repository.NewTranslationRepository(db)
	if err := translationRepo.EnsureSchema(context.Background()); err != nil {
		log.Fatalf("ensure translation schema: %v", err)
	}

	authService := service.NewAuthService(userRepo)
	paperService := service.NewPaperService(db, paperRepo, parseJobRepo)
	agentService, err := agent.NewEinoService(context.Background(), cfg.LLM)
	if err != nil {
		log.Printf("init eino service failed, ai endpoints may be unavailable: %v", err)
		if errors.Is(err, agent.ErrMissingAPIKey) {
			agentService = agent.NewErrorService(err)
		}
	}
	translationService := service.NewTranslationService(paperRepo, translationRepo, agentService)
	knowledgeQAService, err := service.NewKnowledgeQAService(cfg, paperRepo, agentService)
	if err != nil {
		log.Printf("init knowledge qa service failed, rag endpoints may be unavailable: %v", err)
	}

	authHandler := handler.NewAuthHandler(authService)
	paperHandler := handler.NewPaperHandler(knowledgeQAService, paperService, translationService, agentService)

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
			papers.POST("/:id/translate", paperHandler.Translate)
			papers.GET("/:id/translations/latest", paperHandler.LatestTranslation)
			papers.POST("/compare", paperHandler.Compare)
		}
	}
	return r
}
