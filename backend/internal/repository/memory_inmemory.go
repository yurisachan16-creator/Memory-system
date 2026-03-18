package repository

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/yurisachan16-creator/Memory-system/backend/internal/model"
)

type InMemoryMemoryRepository struct {
	mu       sync.RWMutex
	nextID   int64
	memories map[int64]model.Memory
}

func NewInMemoryMemoryRepository() *InMemoryMemoryRepository {
	return &InMemoryMemoryRepository{
		nextID:   1,
		memories: make(map[int64]model.Memory),
	}
}

func (r *InMemoryMemoryRepository) Create(_ context.Context, memory model.Memory) (model.Memory, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	memory.ID = r.nextID
	r.nextID++
	r.memories[memory.ID] = memory.Clone()
	return memory, nil
}

func (r *InMemoryMemoryRepository) Update(_ context.Context, memory model.Memory) (model.Memory, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.memories[memory.ID]; !ok {
		return model.Memory{}, ErrNotFound
	}
	r.memories[memory.ID] = memory.Clone()
	return memory, nil
}

func (r *InMemoryMemoryRepository) GetByID(_ context.Context, id int64) (model.Memory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	memory, ok := r.memories[id]
	if !ok {
		return model.Memory{}, ErrNotFound
	}
	return memory.Clone(), nil
}

func (r *InMemoryMemoryRepository) GetByHash(_ context.Context, userID, contentHash string) (model.Memory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, memory := range r.memories {
		if memory.UserID == userID && memory.ContentHash == contentHash && !memory.IsDeleted {
			return memory.Clone(), nil
		}
	}
	return model.Memory{}, ErrNotFound
}

func (r *InMemoryMemoryRepository) List(_ context.Context, filter ListFilter) ([]model.Memory, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []model.Memory
	for _, memory := range r.memories {
		if memory.IsDeleted || memory.UserID != filter.UserID {
			continue
		}
		if filter.Category != "" && memory.Category != filter.Category {
			continue
		}
		items = append(items, memory.Clone())
	}

	sortMemories(items, filter.SortBy, filter.Order)
	total := len(items)

	start := (filter.Page - 1) * filter.PageSize
	if start >= total {
		return []model.Memory{}, total, nil
	}
	end := start + filter.PageSize
	if end > total {
		end = total
	}
	return items[start:end], total, nil
}

func (r *InMemoryMemoryRepository) ListActiveByUser(_ context.Context, userID string) ([]model.Memory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]model.Memory, 0, len(r.memories))
	for _, memory := range r.memories {
		if memory.UserID == userID && !memory.IsDeleted {
			items = append(items, memory.Clone())
		}
	}
	return items, nil
}

func (r *InMemoryMemoryRepository) SearchCandidates(_ context.Context, userID, query string, limit int) ([]model.Memory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query = strings.ToLower(strings.TrimSpace(query))
	items := make([]model.Memory, 0)
	for _, memory := range r.memories {
		if memory.UserID != userID || memory.IsDeleted {
			continue
		}
		if strings.Contains(strings.ToLower(memory.Content), query) {
			items = append(items, memory.Clone())
		}
	}
	sortMemories(items, "created_at", "desc")
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (r *InMemoryMemoryRepository) SoftDelete(_ context.Context, id int64, userID string, deletedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	memory, ok := r.memories[id]
	if !ok {
		return ErrNotFound
	}
	if memory.UserID != userID {
		return ErrForbidden
	}
	memory.IsDeleted = true
	memory.UpdatedAt = deletedAt
	r.memories[id] = memory
	return nil
}

func sortMemories(items []model.Memory, sortBy, order string) {
	desc := order != "asc"
	sort.Slice(items, func(i, j int) bool {
		left := items[i]
		right := items[j]

		switch sortBy {
		case "importance":
			if left.Importance == right.Importance {
				if desc {
					return left.CreatedAt.After(right.CreatedAt)
				}
				return left.CreatedAt.Before(right.CreatedAt)
			}
			if desc {
				return left.Importance > right.Importance
			}
			return left.Importance < right.Importance
		default:
			if left.CreatedAt.Equal(right.CreatedAt) {
				if desc {
					return left.ID > right.ID
				}
				return left.ID < right.ID
			}
			if desc {
				return left.CreatedAt.After(right.CreatedAt)
			}
			return left.CreatedAt.Before(right.CreatedAt)
		}
	})
}
