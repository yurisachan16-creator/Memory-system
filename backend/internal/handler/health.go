package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"github.com/yurisachan16-creator/Memory-system/backend/internal/response"
)

type Dependencies struct {
	DB    *sql.DB
	Redis *redis.Client
}

type HealthHandler struct {
	db    *sql.DB
	redis *redis.Client
}

func NewHealthHandler(deps Dependencies) *HealthHandler {
	return &HealthHandler{
		db:    deps.DB,
		redis: deps.Redis,
	}
}

// Health godoc
// @Summary Check backend health
// @Description Check service, database and Redis connectivity.
// @Tags health
// @Produce json
// @Success 200 {object} healthResponse
// @Failure 503 {object} errorResponse
// @Router /healthz [get]
func (h *HealthHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	data := gin.H{
		"service":  "up",
		"database": "up",
		"redis":    "up",
	}

	status := http.StatusOK

	if err := h.db.PingContext(ctx); err != nil {
		data["database"] = err.Error()
		status = http.StatusServiceUnavailable
	}

	if err := h.redis.Ping(ctx).Err(); err != nil {
		data["redis"] = err.Error()
		status = http.StatusServiceUnavailable
	}

	if status != http.StatusOK {
		response.Error(c, status, 1001, "dependency check failed")
		return
	}

	response.Success(c, "ok", data)
}

func (h *HealthHandler) Ping(c *gin.Context) {
	response.Success(c, "pong", gin.H{
		"service": "memory-system-backend",
	})
}
