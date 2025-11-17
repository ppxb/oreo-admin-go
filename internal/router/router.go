package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ppxb/oreo-admin-go/internal/middleware"
	"github.com/ppxb/oreo-admin-go/pkg/config"
)

func NewRouter(log *zap.Logger, cfg *config.Config) *gin.Engine {
	if cfg.Server.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.LoggerMiddleware(log))

	public := r.Group("/api/v1")
	{
		public.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})
	}

	return r
}
