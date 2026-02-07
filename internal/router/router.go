package router

import (
	"manjing-ai-go/config"
	"manjing-ai-go/internal/handler"
	"manjing-ai-go/internal/middleware"
	redisclient "manjing-ai-go/pkg/redis"
	swaggerDocs "manjing-ai-go/swagger"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewRouter 构建路由
func NewRouter(cfg *config.Config, authHandler *handler.AuthHandler, resHandler *handler.ResourceHandler, projectHandler *handler.ProjectHandler, chapterHandler *handler.ChapterHandler, emailHandler *handler.EmailHandler, rdb *redisclient.Client) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	if cfg.Storage.Type == "local" {
		r.Static("/storage", cfg.Storage.Local.BaseDir)
	}

	if cfg.Swagger.Enable {
		swaggerDocs.SwaggerInfo.BasePath = "/"
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.InstanceName("swagger")))
	}

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)

			auth.Use(middleware.AuthMiddleware(cfg.JWT, rdb))
			auth.GET("/profile", authHandler.Profile)
			auth.PUT("/password", authHandler.ChangePassword)
			auth.POST("/logout", authHandler.Logout)
		}

		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware(cfg.JWT, rdb))
		{
			users.PUT("/:id/status", authHandler.UpdateStatus)
			users.PUT("/:id/avatar", authHandler.UpdateAvatar)
		}
	}

	v1Public := r.Group("/v1")
	{
		v1Public.POST("/emails/verify-codes", emailHandler.SendVerifyCode)
	}

	v1 := r.Group("/v1")
	v1.Use(middleware.AuthMiddleware(cfg.JWT, rdb))
	{
		v1.POST("/resources", resHandler.Upload)
		v1.GET("/resources", resHandler.List)
		v1.GET("/resources/:id", resHandler.Detail)
		v1.PUT("/resources/:id", resHandler.Update)
		v1.DELETE("/resources/:id", resHandler.Delete)

		v1.POST("/projects", projectHandler.Create)
		v1.GET("/projects", projectHandler.List)
		v1.GET("/projects/:id", projectHandler.Detail)
		v1.PUT("/projects/:id", projectHandler.Update)
		v1.DELETE("/projects/:id", projectHandler.Delete)
		v1.POST("/projects/:id/restore", projectHandler.Restore)
		v1.POST("/projects/:id/archive", projectHandler.Archive)

		v1.POST("/chapters", chapterHandler.Create)
		v1.GET("/chapters", chapterHandler.List)
		v1.GET("/chapters/:id", chapterHandler.Detail)
		v1.PUT("/chapters/:id", chapterHandler.Update)
		v1.DELETE("/chapters/:id", chapterHandler.Delete)
		v1.POST("/chapters/:id/restore", chapterHandler.Restore)
		v1.POST("/chapters/:id/archive", chapterHandler.Archive)
	}

	return r
}
