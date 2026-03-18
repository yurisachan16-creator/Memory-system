package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/yurisachan16-creator/Memory-system/backend/internal/model"
	"github.com/yurisachan16-creator/Memory-system/backend/internal/repository"
)

type countingMemoryRepository struct {
	base            *repository.InMemoryMemoryRepository
	listCalls       int
	activeListCalls int
	searchCalls     int
}

func newCountingMemoryRepository() *countingMemoryRepository {
	return &countingMemoryRepository{base: repository.NewInMemoryMemoryRepository()}
}

func (r *countingMemoryRepository) Create(ctx context.Context, memory model.Memory) (model.Memory, error) {
	return r.base.Create(ctx, memory)
}

func (r *countingMemoryRepository) Update(ctx context.Context, memory model.Memory) (model.Memory, error) {
	return r.base.Update(ctx, memory)
}

func (r *countingMemoryRepository) GetByID(ctx context.Context, id int64) (model.Memory, error) {
	return r.base.GetByID(ctx, id)
}

func (r *countingMemoryRepository) GetByHash(ctx context.Context, userID, contentHash string) (model.Memory, error) {
	return r.base.GetByHash(ctx, userID, contentHash)
}

func (r *countingMemoryRepository) List(ctx context.Context, filter repository.ListFilter) ([]model.Memory, int, error) {
	r.listCalls++
	return r.base.List(ctx, filter)
}

func (r *countingMemoryRepository) ListActiveByUser(ctx context.Context, userID string) ([]model.Memory, error) {
	r.activeListCalls++
	return r.base.ListActiveByUser(ctx, userID)
}

func (r *countingMemoryRepository) SearchCandidates(ctx context.Context, userID, query string, limit int) ([]model.Memory, error) {
	r.searchCalls++
	return r.base.SearchCandidates(ctx, userID, query, limit)
}

func (r *countingMemoryRepository) SoftDelete(ctx context.Context, id int64, userID string, deletedAt time.Time) error {
	return r.base.SoftDelete(ctx, id, userID, deletedAt)
}

func newTestService(now time.Time) *MemoryService {
	svc := NewMemoryService(repository.NewInMemoryMemoryRepository(), repository.NewInMemoryCache())
	svc.now = func() time.Time { return now }
	return svc
}

func TestCreateMemoryDeduplicatesByHash(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	svc := newTestService(now)

	first, merged, err := svc.CreateMemory(context.Background(), model.CreateMemoryRequest{
		UserID:     "u1",
		Content:    "I like coffee",
		Category:   model.CategoryPreference,
		Source:     model.SourceManual,
		Importance: 2,
	})
	if err != nil {
		t.Fatalf("create first memory: %v", err)
	}
	if merged {
		t.Fatalf("expected first create not to merge")
	}

	second, merged, err := svc.CreateMemory(context.Background(), model.CreateMemoryRequest{
		UserID:     "u1",
		Content:    "  i like   coffee ",
		Category:   model.CategoryPreference,
		Source:     model.SourceChat,
		Importance: 5,
	})
	if err != nil {
		t.Fatalf("create duplicate memory: %v", err)
	}
	if !merged {
		t.Fatalf("expected duplicate create to merge")
	}
	if first.ID != second.ID {
		t.Fatalf("expected merged memory id %d, got %d", first.ID, second.ID)
	}
	if second.Importance != 5 {
		t.Fatalf("expected importance to upgrade to 5, got %d", second.Importance)
	}

	list, err := svc.ListMemories(context.Background(), model.ListMemoriesQuery{
		UserID:   "u1",
		SortBy:   "created_at",
		Order:    "desc",
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("list memories: %v", err)
	}
	if list.Total != 1 {
		t.Fatalf("expected one active memory, got %d", list.Total)
	}
}

func TestSearchMemoriesRanksAndCaches(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	svc := newTestService(now)

	seed := []model.CreateMemoryRequest{
		{UserID: "u1", Content: "I love dark roast coffee every morning", Category: model.CategoryPreference, Source: model.SourceManual, Importance: 5},
		{UserID: "u1", Content: "Project alpha goal is to launch next month", Category: model.CategoryGoal, Source: model.SourceManual, Importance: 4},
		{UserID: "u1", Content: "Context note about tea selection", Category: model.CategoryContext, Source: model.SourceManual, Importance: 2},
	}
	for _, item := range seed {
		if _, _, err := svc.CreateMemory(context.Background(), item); err != nil {
			t.Fatalf("seed create: %v", err)
		}
	}

	results, cached, err := svc.SearchMemories(context.Background(), "u1", "coffee", 5)
	if err != nil {
		t.Fatalf("search memories: %v", err)
	}
	if cached {
		t.Fatalf("expected first search not to hit cache")
	}
	if len(results) == 0 || results[0].Memory.Content != "I love dark roast coffee every morning" {
		t.Fatalf("expected coffee memory to rank first, got %+v", results)
	}

	results, cached, err = svc.SearchMemories(context.Background(), "u1", "coffee", 5)
	if err != nil {
		t.Fatalf("search memories from cache: %v", err)
	}
	if !cached {
		t.Fatalf("expected second search to hit cache")
	}
}

func TestSummaryBuildsAndInvalidatesCache(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	svc := newTestService(now)

	fixtures := []model.CreateMemoryRequest{
		{UserID: "u1", Content: "Prefers tea over coffee", Category: model.CategoryPreference, Source: model.SourceManual, Importance: 4},
		{UserID: "u1", Content: "Identity: backend engineer", Category: model.CategoryIdentity, Source: model.SourceManual, Importance: 5},
		{UserID: "u1", Content: "Goal: ship memory API", Category: model.CategoryGoal, Source: model.SourceManual, Importance: 3},
	}
	for _, item := range fixtures {
		if _, _, err := svc.CreateMemory(context.Background(), item); err != nil {
			t.Fatalf("seed create: %v", err)
		}
	}

	summary, cached, err := svc.GetSummary(context.Background(), "u1")
	if err != nil {
		t.Fatalf("get summary: %v", err)
	}
	if cached {
		t.Fatalf("expected first summary not cached")
	}
	if len(summary.Preferences) != 1 || len(summary.Background) != 1 || len(summary.Goals) != 1 {
		t.Fatalf("unexpected summary buckets: %+v", summary)
	}

	_, cached, err = svc.GetSummary(context.Background(), "u1")
	if err != nil {
		t.Fatalf("get cached summary: %v", err)
	}
	if !cached {
		t.Fatalf("expected second summary to be cached")
	}

	list, err := svc.ListMemories(context.Background(), model.ListMemoriesQuery{
		UserID:   "u1",
		SortBy:   "created_at",
		Order:    "desc",
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("list before update: %v", err)
	}

	if _, err := svc.UpdateMemory(context.Background(), list.Items[0].ID, model.UpdateMemoryRequest{
		UserID:     "u1",
		Content:    "Goal: ship memory API this week",
		Category:   model.CategoryGoal,
		Importance: 5,
	}); err != nil {
		t.Fatalf("update memory: %v", err)
	}

	_, cached, err = svc.GetSummary(context.Background(), "u1")
	if err != nil {
		t.Fatalf("get summary after invalidation: %v", err)
	}
	if cached {
		t.Fatalf("expected cache invalidation after write")
	}
}

func TestListMemoriesCachesAndInvalidatesOnWrite(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	repo := newCountingMemoryRepository()
	svc := NewMemoryService(repo, repository.NewInMemoryCache())
	svc.SetNowFunc(func() time.Time { return now })

	for _, item := range []model.CreateMemoryRequest{
		{UserID: "u1", Content: "First memory", Category: model.CategoryContext, Source: model.SourceManual, Importance: 3},
		{UserID: "u1", Content: "Second memory", Category: model.CategoryPreference, Source: model.SourceManual, Importance: 4},
	} {
		if _, _, err := svc.CreateMemory(context.Background(), item); err != nil {
			t.Fatalf("seed create: %v", err)
		}
	}

	query := model.ListMemoriesQuery{
		UserID:   "u1",
		SortBy:   "created_at",
		Order:    "desc",
		Page:     1,
		PageSize: 10,
	}

	first, err := svc.ListMemories(context.Background(), query)
	if err != nil {
		t.Fatalf("list first call: %v", err)
	}
	if repo.listCalls != 1 {
		t.Fatalf("expected first list call to hit repository once, got %d", repo.listCalls)
	}

	second, err := svc.ListMemories(context.Background(), query)
	if err != nil {
		t.Fatalf("list second call: %v", err)
	}
	if repo.listCalls != 1 {
		t.Fatalf("expected second list call to hit cache, got %d repository calls", repo.listCalls)
	}
	if second.Total != first.Total {
		t.Fatalf("expected cached list total %d, got %d", first.Total, second.Total)
	}

	if _, err := svc.UpdateMemory(context.Background(), first.Items[0].ID, model.UpdateMemoryRequest{
		UserID:     "u1",
		Content:    "First memory updated",
		Category:   model.CategoryContext,
		Importance: 5,
	}); err != nil {
		t.Fatalf("update memory: %v", err)
	}

	third, err := svc.ListMemories(context.Background(), query)
	if err != nil {
		t.Fatalf("list after invalidation: %v", err)
	}
	if repo.listCalls != 2 {
		t.Fatalf("expected invalidated list cache to requery repository, got %d calls", repo.listCalls)
	}
	if third.Items[0].Content != "First memory updated" {
		t.Fatalf("expected updated content after invalidation, got %q", third.Items[0].Content)
	}
}

func TestCreateMemoryRejectsTooLongUserID(t *testing.T) {
	svc := newTestService(time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC))

	_, _, err := svc.CreateMemory(context.Background(), model.CreateMemoryRequest{
		UserID:     strings.Repeat("u", 65),
		Content:    "valid content",
		Category:   model.CategoryPreference,
		Source:     model.SourceManual,
		Importance: 3,
	})
	if err != ErrUserIDTooLong {
		t.Fatalf("expected ErrUserIDTooLong, got %v", err)
	}
}

func TestListMemoriesRejectsTooLongUserID(t *testing.T) {
	svc := newTestService(time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC))

	_, err := svc.ListMemories(context.Background(), model.ListMemoriesQuery{
		UserID:   strings.Repeat("u", 65),
		SortBy:   "created_at",
		Order:    "desc",
		Page:     1,
		PageSize: 10,
	})
	if err != ErrUserIDTooLong {
		t.Fatalf("expected ErrUserIDTooLong, got %v", err)
	}
}
