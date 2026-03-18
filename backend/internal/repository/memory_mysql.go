package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/yurisachan16-creator/Memory-system/backend/internal/model"
)

type MySQLMemoryRepository struct {
	db *sql.DB
}

func NewMySQLMemoryRepository(db *sql.DB) *MySQLMemoryRepository {
	return &MySQLMemoryRepository{db: db}
}

func (r *MySQLMemoryRepository) Create(ctx context.Context, memory model.Memory) (model.Memory, error) {
	query := `
		INSERT INTO memories (
			user_id, content, category, source, importance, content_hash, is_deleted, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(
		ctx,
		query,
		memory.UserID,
		memory.Content,
		memory.Category,
		memory.Source,
		memory.Importance,
		memory.ContentHash,
		memory.IsDeleted,
		memory.CreatedAt,
		memory.UpdatedAt,
	)
	if err != nil {
		return model.Memory{}, fmt.Errorf("insert memory: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.Memory{}, fmt.Errorf("read insert id: %w", err)
	}
	memory.ID = id
	return memory, nil
}

func (r *MySQLMemoryRepository) Update(ctx context.Context, memory model.Memory) (model.Memory, error) {
	query := `
		UPDATE memories
		SET content = ?, category = ?, importance = ?, content_hash = ?, updated_at = ?
		WHERE id = ? AND is_deleted = 0
	`
	result, err := r.db.ExecContext(
		ctx,
		query,
		memory.Content,
		memory.Category,
		memory.Importance,
		memory.ContentHash,
		memory.UpdatedAt,
		memory.ID,
	)
	if err != nil {
		return model.Memory{}, fmt.Errorf("update memory: %w", err)
	}
	rows, err := result.RowsAffected()
	if err == nil && rows == 0 {
		return model.Memory{}, ErrNotFound
	}
	return memory, nil
}

func (r *MySQLMemoryRepository) GetByID(ctx context.Context, id int64) (model.Memory, error) {
	query := `
		SELECT id, user_id, content, category, source, importance, content_hash, is_deleted, created_at, updated_at
		FROM memories
		WHERE id = ?
	`
	var memory model.Memory
	var category string
	var source string
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&memory.ID,
		&memory.UserID,
		&memory.Content,
		&category,
		&source,
		&memory.Importance,
		&memory.ContentHash,
		&memory.IsDeleted,
		&memory.CreatedAt,
		&memory.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return model.Memory{}, ErrNotFound
		}
		return model.Memory{}, fmt.Errorf("get memory by id: %w", err)
	}
	memory.Category = model.Category(category)
	memory.Source = model.Source(source)
	return memory, nil
}

func (r *MySQLMemoryRepository) GetByHash(ctx context.Context, userID, contentHash string) (model.Memory, error) {
	query := `
		SELECT id, user_id, content, category, source, importance, content_hash, is_deleted, created_at, updated_at
		FROM memories
		WHERE user_id = ? AND content_hash = ? AND is_deleted = 0
		ORDER BY updated_at DESC
		LIMIT 1
	`
	var memory model.Memory
	var category string
	var source string
	if err := r.db.QueryRowContext(ctx, query, userID, contentHash).Scan(
		&memory.ID,
		&memory.UserID,
		&memory.Content,
		&category,
		&source,
		&memory.Importance,
		&memory.ContentHash,
		&memory.IsDeleted,
		&memory.CreatedAt,
		&memory.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return model.Memory{}, ErrNotFound
		}
		return model.Memory{}, fmt.Errorf("get memory by hash: %w", err)
	}
	memory.Category = model.Category(category)
	memory.Source = model.Source(source)
	return memory, nil
}

func (r *MySQLMemoryRepository) List(ctx context.Context, filter ListFilter) ([]model.Memory, int, error) {
	countQuery := `SELECT COUNT(*) FROM memories WHERE user_id = ? AND is_deleted = 0`
	countArgs := []interface{}{filter.UserID}
	if filter.Category != "" {
		countQuery += ` AND category = ?`
		countArgs = append(countArgs, filter.Category)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count memories: %w", err)
	}

	orderBy := "created_at"
	if filter.SortBy == "importance" {
		orderBy = "importance"
	}
	order := "DESC"
	if strings.EqualFold(filter.Order, "asc") {
		order = "ASC"
	}

	query := `
		SELECT id, user_id, content, category, source, importance, content_hash, is_deleted, created_at, updated_at
		FROM memories
		WHERE user_id = ? AND is_deleted = 0
	`
	args := []interface{}{filter.UserID}
	if filter.Category != "" {
		query += ` AND category = ?`
		args = append(args, filter.Category)
	}
	query += fmt.Sprintf(` ORDER BY %s %s, id DESC LIMIT ? OFFSET ?`, orderBy, order)
	args = append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list memories: %w", err)
	}
	defer rows.Close()

	items, err := scanMemories(rows)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *MySQLMemoryRepository) ListActiveByUser(ctx context.Context, userID string) ([]model.Memory, error) {
	query := `
		SELECT id, user_id, content, category, source, importance, content_hash, is_deleted, created_at, updated_at
		FROM memories
		WHERE user_id = ? AND is_deleted = 0
		ORDER BY created_at DESC, id DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list active memories: %w", err)
	}
	defer rows.Close()
	return scanMemories(rows)
}

func (r *MySQLMemoryRepository) SearchCandidates(ctx context.Context, userID, query string, limit int) ([]model.Memory, error) {
	if limit <= 0 {
		limit = 20
	}

	queryText := strings.TrimSpace(query)
	likePattern := "%" + queryText + "%"
	sqlQuery := `
		SELECT id, user_id, content, category, source, importance, content_hash, is_deleted, created_at, updated_at
		FROM memories
		WHERE user_id = ? AND is_deleted = 0
		  AND (
			MATCH(content) AGAINST(? IN NATURAL LANGUAGE MODE)
			OR content LIKE ?
		  )
		ORDER BY
		  MATCH(content) AGAINST(? IN NATURAL LANGUAGE MODE) DESC,
		  importance DESC,
		  created_at DESC,
		  id DESC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, sqlQuery, userID, queryText, likePattern, queryText, limit)
	if err != nil {
		return nil, fmt.Errorf("search memories: %w", err)
	}
	defer rows.Close()
	return scanMemories(rows)
}

func (r *MySQLMemoryRepository) SoftDelete(ctx context.Context, id int64, userID string, deletedAt time.Time) error {
	query := `
		UPDATE memories
		SET is_deleted = 1, updated_at = ?
		WHERE id = ? AND user_id = ? AND is_deleted = 0
	`
	result, err := r.db.ExecContext(ctx, query, deletedAt, id, userID)
	if err != nil {
		return fmt.Errorf("soft delete memory: %w", err)
	}
	rows, err := result.RowsAffected()
	if err == nil && rows > 0 {
		return nil
	}

	memory, getErr := r.GetByID(ctx, id)
	if getErr != nil {
		return getErr
	}
	if memory.UserID != userID {
		return ErrForbidden
	}
	if memory.IsDeleted {
		return ErrNotFound
	}
	return ErrNotFound
}

func scanMemories(rows *sql.Rows) ([]model.Memory, error) {
	items := make([]model.Memory, 0)
	for rows.Next() {
		var memory model.Memory
		var category string
		var source string
		if err := rows.Scan(
			&memory.ID,
			&memory.UserID,
			&memory.Content,
			&category,
			&source,
			&memory.Importance,
			&memory.ContentHash,
			&memory.IsDeleted,
			&memory.CreatedAt,
			&memory.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan memory: %w", err)
		}
		memory.Category = model.Category(category)
		memory.Source = model.Source(source)
		items = append(items, memory)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate memories: %w", err)
	}
	return items, nil
}
