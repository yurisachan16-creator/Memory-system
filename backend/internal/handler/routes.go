package handler

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/yurisachan16-creator/Memory-system/backend/internal/repository"
	"github.com/yurisachan16-creator/Memory-system/backend/internal/service"
)

func RegisterRoutes(router *gin.Engine, deps Dependencies) {
	healthHandler := NewHealthHandler(deps)
	memoryRepo := repository.NewMySQLMemoryRepository(deps.DB)
	cache := repository.NewRedisCache(deps.Redis)
	memoryService := service.NewMemoryService(memoryRepo, cache)

	router.GET("/healthz", healthHandler.Health)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api/v1")
	{
		api.GET("/ping", healthHandler.Ping)
		RegisterMemoryRoutes(api, memoryService)
	}
}
