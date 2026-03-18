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

// Create godoc
// @Summary Create a memory
// @Description Create a memory record. Duplicate content for the same user merges into the existing memory and keeps the higher importance.
// @Tags memories
// @Accept json
// @Produce json
// @Param request body model.CreateMemoryRequest true "Create memory payload"
// @Success 201 {object} createMemoryResponse
// @Success 200 {object} createMemoryResponse "Merged duplicate memory"
// @Failure 400 {object} errorResponse
// @Failure 409 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /memories [post]
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

// List godoc
// @Summary List memories
// @Description List memories for a user with optional category filter, sorting and pagination.
// @Tags memories
// @Accept json
// @Produce json
// @Param user_id query string true "User ID"
// @Param category query string false "Category" Enums(preference, identity, goal, context)
// @Param sort_by query string false "Sort field" Enums(importance, created_at) default(created_at)
// @Param order query string false "Sort order" Enums(asc, desc) default(desc)
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} listMemoriesResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /memories [get]
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

// Update godoc
// @Summary Update a memory
// @Description Update content, category and importance for a memory that belongs to the specified user.
// @Tags memories
// @Accept json
// @Produce json
// @Param id path int true "Memory ID"
// @Param request body model.UpdateMemoryRequest true "Update memory payload"
// @Success 200 {object} createMemoryResponse
// @Failure 400 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 409 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /memories/{id} [put]
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

// Delete godoc
// @Summary Delete a memory
// @Description Soft delete a memory that belongs to the specified user.
// @Tags memories
// @Accept json
// @Produce json
// @Param id path int true "Memory ID"
// @Param user_id query string true "User ID"
// @Success 200 {object} deleteMemoryResponse
// @Failure 400 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /memories/{id} [delete]
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

// Search godoc
// @Summary Search memories
// @Description Search memories for a user and return the top ranked results with cache status.
// @Tags memories
// @Accept json
// @Produce json
// @Param user_id query string true "User ID"
// @Param query query string true "Search query"
// @Param limit query int false "Result limit (3-5)" default(5)
// @Success 200 {object} searchMemoriesResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /memories/search [get]
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

// Summary godoc
// @Summary Summarize memories
// @Description Build a memory summary for a user including preferences, goals, background and recent memories.
// @Tags memories
// @Accept json
// @Produce json
// @Param user_id query string true "User ID"
// @Success 200 {object} summaryResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /memories/summary [get]
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
		errors.Is(err, service.ErrUserIDTooLong),
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
	case errors.Is(err, service.ErrCreateInProgress):
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
