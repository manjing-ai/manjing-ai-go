package main

import (
	"context"
	"flag"

	"manjing-ai-go/config"
	"manjing-ai-go/internal/handler"
	"manjing-ai-go/internal/repository"
	"manjing-ai-go/internal/router"
	"manjing-ai-go/internal/service"
	"manjing-ai-go/pkg/logger"
	redisclient "manjing-ai-go/pkg/redis"
	"manjing-ai-go/pkg/storage"

	"github.com/gin-gonic/gin"
)

// @title Manjing AI API
// @version 1.0
// @description Manjing AI backend API.
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @BasePath /

func main() {
	logger.Init()
	configPath := flag.String("config", "", "config file path")
	flag.Parse()
	cfg := config.MustLoadWithPath(*configPath)
	gin.SetMode(cfg.App.Mode)

	db, err := repository.InitDB(cfg.DB.DSN)
	if err != nil {
		panic(err)
	}

	rdb := redisclient.New(cfg.Redis)
	if err := rdb.Ping(context.Background()); err != nil {
		// Redis 可选，初始化失败不阻断启动
		rdb = nil
	}

	userRepo := repository.NewUserRepo(db)
	authSvc := service.NewAuthService(userRepo, cfg.JWT, rdb)
	authHandler := handler.NewAuthHandler(authSvc)

	var storageSvc storage.Service
	switch cfg.Storage.Type {
	case "cos":
		storageSvc = storage.NewCOSStorage(cfg.Storage.COS)
	default:
		storageSvc = storage.NewLocalStorage(cfg.Storage.Local)
	}

	resRepo := repository.NewResourceRepo(db)
	resSvc := service.NewResourceService(resRepo, storageSvc, cfg.Storage)
	resHandler := handler.NewResourceHandler(resSvc)

	projectRepo := repository.NewProjectRepo(db)
	projectSvc := service.NewProjectService(projectRepo)
	projectHandler := handler.NewProjectHandler(projectSvc)

	chapterRepo := repository.NewChapterRepo(db)
	chapterSvc := service.NewChapterService(chapterRepo, projectRepo)
	chapterHandler := handler.NewChapterHandler(chapterSvc)

	r := router.NewRouter(cfg, authHandler, resHandler, projectHandler, chapterHandler, rdb)
	logger.L().Info("api listening on ", cfg.App.Addr)
	_ = r.Run(cfg.App.Addr)
}
