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
func NewRouter(cfg *config.Config, authHandler *handler.AuthHandler, rdb *redisclient.Client) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

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

	return r
}
