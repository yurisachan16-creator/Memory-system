package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/yurisachan16-creator/Memory-system/backend/internal/repository"
	"github.com/yurisachan16-creator/Memory-system/backend/internal/service"
)

func RegisterRoutes(router *gin.Engine, deps Dependencies) {
	healthHandler := NewHealthHandler(deps)
	memoryRepo := repository.NewMySQLMemoryRepository(deps.DB)
	cache := repository.NewRedisCache(deps.Redis)
	memoryService := service.NewMemoryService(memoryRepo, cache)

	router.GET("/healthz", healthHandler.Health)

	api := router.Group("/api/v1")
	{
		api.GET("/ping", healthHandler.Ping)
		RegisterMemoryRoutes(api, memoryService)
	}
}
