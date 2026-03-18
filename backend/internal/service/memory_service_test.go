package service

import (
	"context"
	"testing"
	"time"

	"github.com/yurisachan16-creator/Memory-system/backend/internal/model"
	"github.com/yurisachan16-creator/Memory-system/backend/internal/repository"
)

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
