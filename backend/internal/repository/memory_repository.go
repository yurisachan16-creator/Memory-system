package repository

import (
	"context"
	"errors"
	"time"

	"github.com/yurisachan16-creator/Memory-system/backend/internal/model"
)

var (
	ErrNotFound  = errors.New("memory not found")
	ErrForbidden = errors.New("memory does not belong to user")
)

type ListFilter struct {
	UserID   string
	Category model.Category
	SortBy   string
	Order    string
	Page     int
	PageSize int
}

type MemoryRepository interface {
	Create(ctx context.Context, memory model.Memory) (model.Memory, error)
	Update(ctx context.Context, memory model.Memory) (model.Memory, error)
	GetByID(ctx context.Context, id int64) (model.Memory, error)
	GetByHash(ctx context.Context, userID, contentHash string) (model.Memory, error)
	List(ctx context.Context, filter ListFilter) ([]model.Memory, int, error)
	ListActiveByUser(ctx context.Context, userID string) ([]model.Memory, error)
	SearchCandidates(ctx context.Context, userID, query string, limit int) ([]model.Memory, error)
	SoftDelete(ctx context.Context, id int64, userID string, deletedAt time.Time) error
}
