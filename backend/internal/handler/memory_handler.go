package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/yurisachan16-creator/Memory-system/backend/internal/model"
	"github.com/yurisachan16-creator/Memory-system/backend/internal/repository"
	"github.com/yurisachan16-creator/Memory-system/backend/internal/response"
	"github.com/yurisachan16-creator/Memory-system/backend/internal/service"
)

type MemoryHandler struct {
	service *service.MemoryService
}

func NewMemoryHandler(service *service.MemoryService) *MemoryHandler {
	return &MemoryHandler{service: service}
}

func RegisterMemoryRoutes(api *gin.RouterGroup, service *service.MemoryService) {
	handler := NewMemoryHandler(service)
	api.POST("/memories", handler.Create)
	api.GET("/memories", handler.List)
	api.PUT("/memories/:id", handler.Update)
	api.DELETE("/memories/:id", handler.Delete)
	api.GET("/memories/search", handler.Search)
	api.GET("/memories/summary", handler.Summary)
}

func (h *MemoryHandler) Create(c *gin.Context) {
	var req model.CreateMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "invalid json body")
		return
	}

	memory, merged, err := h.service.CreateMemory(c.Request.Context(), req)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	if merged {
		response.Success(c, "merged duplicate memory", memory)
		return
	}
	response.SuccessWithStatus(c, http.StatusCreated, "created", memory)
}

func (h *MemoryHandler) List(c *gin.Context) {
	query := model.ListMemoriesQuery{
		UserID:   strings.TrimSpace(c.Query("user_id")),
		Category: model.Category(strings.TrimSpace(c.Query("category"))),
		SortBy:   strings.TrimSpace(c.DefaultQuery("sort_by", "created_at")),
		Order:    strings.TrimSpace(c.DefaultQuery("order", "desc")),
		Page:     parseIntOrDefault(c.Query("page"), 1),
		PageSize: parseIntOrDefault(c.Query("page_size"), 10),
	}

	result, err := h.service.ListMemories(c.Request.Context(), query)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, "ok", result)
}

func (h *MemoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusNotFound, http.StatusNotFound, "invalid memory id")
		return
	}

	var req model.UpdateMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "invalid json body")
		return
	}

	memory, err := h.service.UpdateMemory(c.Request.Context(), id, req)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, "updated", memory)
}

func (h *MemoryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusNotFound, http.StatusNotFound, "invalid memory id")
		return
	}

	if err := h.service.DeleteMemory(c.Request.Context(), id, strings.TrimSpace(c.Query("user_id"))); err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, "deleted", gin.H{"id": id})
}

func (h *MemoryHandler) Search(c *gin.Context) {
	results, cached, err := h.service.SearchMemories(
		c.Request.Context(),
		strings.TrimSpace(c.Query("user_id")),
		strings.TrimSpace(c.Query("query")),
		parseIntOrDefault(c.Query("limit"), 5),
	)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, "ok", gin.H{
		"items":  results,
		"count":  len(results),
		"cached": cached,
	})
}

func (h *MemoryHandler) Summary(c *gin.Context) {
	summary, cached, err := h.service.GetSummary(c.Request.Context(), strings.TrimSpace(c.Query("user_id")))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, "ok", gin.H{
		"summary": summary,
		"cached":  cached,
	})
}

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidUserID),
		errors.Is(err, service.ErrInvalidContent),
		errors.Is(err, service.ErrInvalidCategory),
		errors.Is(err, service.ErrInvalidSource),
		errors.Is(err, service.ErrInvalidImportance),
		errors.Is(err, service.ErrInvalidSortBy),
		errors.Is(err, service.ErrInvalidOrder),
		errors.Is(err, service.ErrInvalidPage),
		errors.Is(err, service.ErrInvalidPageSize),
		errors.Is(err, service.ErrInvalidQuery):
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrDuplicateUpdate):
		response.Error(c, http.StatusConflict, http.StatusConflict, err.Error())
	case errors.Is(err, repository.ErrForbidden):
		response.Error(c, http.StatusForbidden, http.StatusForbidden, err.Error())
	case errors.Is(err, repository.ErrNotFound):
		response.Error(c, http.StatusNotFound, http.StatusNotFound, err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "internal server error")
	}
}

func parseIntOrDefault(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
