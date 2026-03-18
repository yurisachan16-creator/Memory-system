package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/yurisachan16-creator/Memory-system/backend/internal/model"
	"github.com/yurisachan16-creator/Memory-system/backend/internal/repository"
)

const (
	searchCacheTTL  = 5 * time.Minute
	summaryCacheTTL = 10 * time.Minute
)

var (
	ErrInvalidUserID     = errors.New("user_id is required")
	ErrInvalidContent    = errors.New("content must not be empty")
	ErrInvalidCategory   = errors.New("category is invalid")
	ErrInvalidSource     = errors.New("source is invalid")
	ErrInvalidSortBy     = errors.New("sort_by must be importance or created_at")
	ErrInvalidOrder      = errors.New("order must be asc or desc")
	ErrInvalidPage       = errors.New("page must be greater than 0")
	ErrInvalidPageSize   = errors.New("page_size must be between 1 and 100")
	ErrInvalidImportance = errors.New("importance must be between 1 and 5")
	ErrInvalidQuery      = errors.New("query is required")
	ErrDuplicateUpdate   = errors.New("another memory with the same content already exists")
)

type MemoryService struct {
	repo  repository.MemoryRepository
	cache repository.Cache
	now   func() time.Time
}

func NewMemoryService(repo repository.MemoryRepository, cache repository.Cache) *MemoryService {
	return &MemoryService{
		repo:  repo,
		cache: cache,
		now:   time.Now,
	}
}

func (s *MemoryService) SetNowFunc(now func() time.Time) {
	if now == nil {
		s.now = time.Now
		return
	}
	s.now = now
}

func (s *MemoryService) CreateMemory(ctx context.Context, req model.CreateMemoryRequest) (model.Memory, bool, error) {
	if err := validateCreateRequest(req); err != nil {
		return model.Memory{}, false, err
	}

	now := s.now().UTC()
	req.Content = model.NormalizeContent(req.Content)
	contentHash := model.HashContent(req.Content)

	existing, err := s.repo.GetByHash(ctx, req.UserID, contentHash)
	if err == nil {
		merged := existing.Clone()
		if req.Importance > merged.Importance {
			merged.Importance = req.Importance
		}
		merged.UpdatedAt = now
		updated, updateErr := s.repo.Update(ctx, merged)
		if updateErr != nil {
			return model.Memory{}, false, updateErr
		}
		s.invalidateUserCaches(req.UserID)
		return updated, true, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return model.Memory{}, false, err
	}

	memory := model.Memory{
		UserID:      req.UserID,
		Content:     req.Content,
		Category:    req.Category,
		Source:      req.Source,
		Importance:  req.Importance,
		ContentHash: contentHash,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	created, err := s.repo.Create(ctx, memory)
	if err != nil {
		return model.Memory{}, false, err
	}
	s.invalidateUserCaches(req.UserID)
	return created, false, nil
}

func (s *MemoryService) ListMemories(ctx context.Context, query model.ListMemoriesQuery) (model.ListMemoriesResult, error) {
	if query.SortBy == "" {
		query.SortBy = "created_at"
	}
	if query.Order == "" {
		query.Order = "desc"
	}
	if err := validateListQuery(query); err != nil {
		return model.ListMemoriesResult{}, err
	}
	items, total, err := s.repo.List(ctx, repository.ListFilter{
		UserID:   query.UserID,
		Category: query.Category,
		SortBy:   query.SortBy,
		Order:    query.Order,
		Page:     query.Page,
		PageSize: query.PageSize,
	})
	if err != nil {
		return model.ListMemoriesResult{}, err
	}
	return model.ListMemoriesResult{
		Items:    items,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

func (s *MemoryService) UpdateMemory(ctx context.Context, id int64, req model.UpdateMemoryRequest) (model.Memory, error) {
	if err := validateUpdateRequest(req); err != nil {
		return model.Memory{}, err
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return model.Memory{}, err
	}
	if existing.IsDeleted {
		return model.Memory{}, repository.ErrNotFound
	}
	if existing.UserID != req.UserID {
		return model.Memory{}, repository.ErrForbidden
	}

	normalizedContent := model.NormalizeContent(req.Content)
	newHash := model.HashContent(normalizedContent)
	if newHash != existing.ContentHash {
		duplicate, dupErr := s.repo.GetByHash(ctx, req.UserID, newHash)
		if dupErr == nil && duplicate.ID != existing.ID {
			return model.Memory{}, ErrDuplicateUpdate
		}
		if dupErr != nil && !errors.Is(dupErr, repository.ErrNotFound) {
			return model.Memory{}, dupErr
		}
	}

	existing.Content = normalizedContent
	existing.Category = req.Category
	existing.Importance = req.Importance
	existing.ContentHash = newHash
	existing.UpdatedAt = s.now().UTC()

	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		return model.Memory{}, err
	}
	s.invalidateUserCaches(req.UserID)
	return updated, nil
}

func (s *MemoryService) DeleteMemory(ctx context.Context, id int64, userID string) error {
	if strings.TrimSpace(userID) == "" {
		return ErrInvalidUserID
	}
	if err := s.repo.SoftDelete(ctx, id, userID, s.now().UTC()); err != nil {
		return err
	}
	s.invalidateUserCaches(userID)
	return nil
}

func (s *MemoryService) SearchMemories(ctx context.Context, userID, query string, limit int) ([]model.SearchMemoryResult, bool, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, false, ErrInvalidUserID
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, false, ErrInvalidQuery
	}
	if limit <= 0 {
		limit = 5
	}
	if limit < 3 {
		limit = 3
	}
	if limit > 5 {
		limit = 5
	}

	cacheKey := fmt.Sprintf("memories:search:%s:%s", userID, model.HashContent(query))
	if payload, ok := s.cache.Get(cacheKey); ok {
		var cached []model.SearchMemoryResult
		if err := json.Unmarshal(payload, &cached); err == nil {
			return cached, true, nil
		}
	}

	memories, err := s.repo.SearchCandidates(ctx, userID, query, limit*4)
	if err != nil {
		return nil, false, err
	}

	queryTokens := tokenize(query)
	results := make([]model.SearchMemoryResult, 0, len(memories))
	now := s.now().UTC()
	queryLower := strings.ToLower(query)
	for _, memory := range memories {
		relevance, matchedTerms := scoreRelevance(memory.Content, queryLower, queryTokens)
		if relevance == 0 {
			continue
		}
		recency := scoreRecency(memory.CreatedAt, now)
		importanceScore := float64(memory.Importance) / 5.0
		finalScore := (relevance * 0.5) + (importanceScore * 0.3) + (recency * 0.2)
		results = append(results, model.SearchMemoryResult{
			Memory:         memory,
			RelevanceScore: roundScore(relevance),
			RecencyScore:   roundScore(recency),
			FinalScore:     roundScore(finalScore),
			MatchedTerms:   matchedTerms,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].FinalScore == results[j].FinalScore {
			return results[i].Memory.UpdatedAt.After(results[j].Memory.UpdatedAt)
		}
		return results[i].FinalScore > results[j].FinalScore
	})

	if len(results) > limit {
		results = results[:limit]
	}

	if payload, err := json.Marshal(results); err == nil {
		s.cache.Set(cacheKey, payload, searchCacheTTL)
	}
	return results, false, nil
}

func (s *MemoryService) GetSummary(ctx context.Context, userID string) (model.SummaryResult, bool, error) {
	if strings.TrimSpace(userID) == "" {
		return model.SummaryResult{}, false, ErrInvalidUserID
	}
	cacheKey := fmt.Sprintf("memories:summary:%s", userID)
	if payload, ok := s.cache.Get(cacheKey); ok {
		var cached model.SummaryResult
		if err := json.Unmarshal(payload, &cached); err == nil {
			return cached, true, nil
		}
	}

	memories, err := s.repo.ListActiveByUser(ctx, userID)
	if err != nil {
		return model.SummaryResult{}, false, err
	}

	now := s.now().UTC()
	result := model.SummaryResult{
		Preferences: summarizePreferences(memories),
		Goals:       summarizeGoals(memories),
		Background:  summarizeBackground(memories),
		Recent:      summarizeRecent(memories, now),
	}
	if payload, err := json.Marshal(result); err == nil {
		s.cache.Set(cacheKey, payload, summaryCacheTTL)
	}
	return result, false, nil
}

func (s *MemoryService) invalidateUserCaches(userID string) {
	s.cache.DeleteByPrefix(fmt.Sprintf("memories:search:%s:", userID))
	s.cache.Delete(fmt.Sprintf("memories:summary:%s", userID))
}

func validateCreateRequest(req model.CreateMemoryRequest) error {
	if strings.TrimSpace(req.UserID) == "" {
		return ErrInvalidUserID
	}
	if model.NormalizeContent(req.Content) == "" {
		return ErrInvalidContent
	}
	if !req.Category.Valid() {
		return ErrInvalidCategory
	}
	if !req.Source.Valid() {
		return ErrInvalidSource
	}
	if req.Importance < 1 || req.Importance > 5 {
		return ErrInvalidImportance
	}
	return nil
}

func validateUpdateRequest(req model.UpdateMemoryRequest) error {
	if strings.TrimSpace(req.UserID) == "" {
		return ErrInvalidUserID
	}
	if model.NormalizeContent(req.Content) == "" {
		return ErrInvalidContent
	}
	if !req.Category.Valid() {
		return ErrInvalidCategory
	}
	if req.Importance < 1 || req.Importance > 5 {
		return ErrInvalidImportance
	}
	return nil
}

func validateListQuery(query model.ListMemoriesQuery) error {
	if strings.TrimSpace(query.UserID) == "" {
		return ErrInvalidUserID
	}
	if query.Category != "" && !query.Category.Valid() {
		return ErrInvalidCategory
	}
	if query.SortBy != "importance" && query.SortBy != "created_at" {
		return ErrInvalidSortBy
	}
	if query.Order != "asc" && query.Order != "desc" {
		return ErrInvalidOrder
	}
	if query.Page <= 0 {
		return ErrInvalidPage
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		return ErrInvalidPageSize
	}
	return nil
}

func tokenize(query string) []string {
	query = strings.ToLower(query)
	fields := strings.FieldsFunc(query, func(r rune) bool {
		switch {
		case r >= 'a' && r <= 'z':
			return false
		case r >= '0' && r <= '9':
			return false
		case r >= 0x4E00 && r <= 0x9FFF:
			return false
		default:
			return true
		}
	})
	if len(fields) == 0 {
		return []string{query}
	}
	seen := make(map[string]struct{}, len(fields))
	tokens := make([]string, 0, len(fields))
	for _, field := range fields {
		if field == "" {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		tokens = append(tokens, field)
	}
	return tokens
}

func scoreRelevance(content, queryLower string, queryTokens []string) (float64, []string) {
	contentLower := strings.ToLower(content)
	matched := make([]string, 0, len(queryTokens))
	for _, token := range queryTokens {
		if strings.Contains(contentLower, token) {
			matched = append(matched, token)
		}
	}

	fulltextScore := 0.0
	if len(queryTokens) > 0 {
		fulltextScore = float64(len(matched)) / float64(len(queryTokens))
	}

	likeScore := 0.0
	if strings.Contains(contentLower, queryLower) {
		likeScore = 1.0
	}

	relevance := math.Max(fulltextScore, likeScore*0.85)
	return roundScore(relevance), matched
}

func scoreRecency(createdAt, now time.Time) float64 {
	days := now.Sub(createdAt).Hours() / 24
	if days <= 0 {
		return 1
	}
	decayWindow := 30.0
	score := 1 - (days / decayWindow)
	if score < 0 {
		return 0
	}
	return roundScore(score)
}

func summarizePreferences(memories []model.Memory) []model.Memory {
	items := make([]model.Memory, 0)
	for _, memory := range memories {
		if memory.Category == model.CategoryPreference && memory.Importance >= 3 {
			items = append(items, memory)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Importance == items[j].Importance {
			return items[i].UpdatedAt.After(items[j].UpdatedAt)
		}
		return items[i].Importance > items[j].Importance
	})
	return items
}

func summarizeGoals(memories []model.Memory) []model.Memory {
	var latest *model.Memory
	for _, memory := range memories {
		if memory.Category != model.CategoryGoal {
			continue
		}
		if latest == nil || memory.CreatedAt.After(latest.CreatedAt) {
			current := memory
			latest = &current
		}
	}
	if latest == nil {
		return []model.Memory{}
	}
	return []model.Memory{latest.Clone()}
}

func summarizeBackground(memories []model.Memory) []model.Memory {
	var selected *model.Memory
	for _, memory := range memories {
		if memory.Category != model.CategoryIdentity {
			continue
		}
		if selected == nil ||
			memory.Importance > selected.Importance ||
			(memory.Importance == selected.Importance && memory.UpdatedAt.After(selected.UpdatedAt)) {
			current := memory
			selected = &current
		}
	}
	if selected == nil {
		return []model.Memory{}
	}
	return []model.Memory{selected.Clone()}
}

func summarizeRecent(memories []model.Memory, now time.Time) []model.Memory {
	items := make([]model.Memory, 0)
	for _, memory := range memories {
		if now.Sub(memory.CreatedAt) <= 7*24*time.Hour {
			items = append(items, memory)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items
}

func roundScore(value float64) float64 {
	rounded, _ := strconv.ParseFloat(fmt.Sprintf("%.4f", value), 64)
	return rounded
}
