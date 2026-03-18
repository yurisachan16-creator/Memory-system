package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/yurisachan16-creator/Memory-system/backend/internal/model"
)

func TestMySQLMemoryRepositoryList(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewMySQLMemoryRepository(db)
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM memories WHERE user_id = ? AND is_deleted = 0 AND category = ?`)).
		WithArgs("u1", model.CategoryPreference).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "content", "category", "source", "importance", "content_hash", "is_deleted", "created_at", "updated_at",
	}).
		AddRow(2, "u1", "Second", "preference", "manual", 5, "hash-2", false, now, now).
		AddRow(1, "u1", "First", "preference", "chat", 3, "hash-1", false, now.Add(-time.Hour), now.Add(-time.Hour))

	mock.ExpectQuery(`(?s)SELECT id, user_id, content, category, source, importance, content_hash, is_deleted, created_at, updated_at\s+FROM memories\s+WHERE user_id = \? AND is_deleted = 0\s+AND category = \? ORDER BY importance DESC, id DESC LIMIT \? OFFSET \?`).
		WithArgs("u1", model.CategoryPreference, 10, 0).
		WillReturnRows(rows)

	items, total, err := repo.List(context.Background(), ListFilter{
		UserID:   "u1",
		Category: model.CategoryPreference,
		SortBy:   "importance",
		Order:    "desc",
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("list memories: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}
	if len(items) != 2 || items[0].ID != 2 {
		t.Fatalf("unexpected list result: %+v", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestMySQLMemoryRepositorySearchCandidates(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewMySQLMemoryRepository(db)
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "content", "category", "source", "importance", "content_hash", "is_deleted", "created_at", "updated_at",
	}).AddRow(7, "u1", "coffee goal", "goal", "manual", 4, "hash-7", false, now, now)

	mock.ExpectQuery(`(?s)SELECT id, user_id, content, category, source, importance, content_hash, is_deleted, created_at, updated_at\s+FROM memories\s+WHERE user_id = \? AND is_deleted = 0\s+AND \(\s+MATCH\(content\) AGAINST\(\? IN NATURAL LANGUAGE MODE\)\s+OR content LIKE \?\s+\)\s+ORDER BY\s+MATCH\(content\) AGAINST\(\? IN NATURAL LANGUAGE MODE\) DESC,\s+importance DESC,\s+created_at DESC,\s+id DESC\s+LIMIT \?`).
		WithArgs("u1", "coffee", "%coffee%", "coffee", 4).
		WillReturnRows(rows)

	items, err := repo.SearchCandidates(context.Background(), "u1", "coffee", 4)
	if err != nil {
		t.Fatalf("search candidates: %v", err)
	}
	if len(items) != 1 || items[0].ID != 7 {
		t.Fatalf("unexpected search result: %+v", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestMySQLMemoryRepositorySoftDeleteForbidden(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewMySQLMemoryRepository(db)
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)

	mock.ExpectExec(`(?s)UPDATE memories\s+SET is_deleted = 1, updated_at = \?\s+WHERE id = \? AND user_id = \? AND is_deleted = 0`).
		WithArgs(now, int64(3), "intruder").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectQuery(`(?s)SELECT id, user_id, content, category, source, importance, content_hash, is_deleted, created_at, updated_at\s+FROM memories\s+WHERE id = \?`).
		WithArgs(int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "content", "category", "source", "importance", "content_hash", "is_deleted", "created_at", "updated_at",
		}).AddRow(3, "owner", "Original", "context", "manual", 3, "hash-3", false, now, now))

	err = repo.SoftDelete(context.Background(), 3, "intruder", now)
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
