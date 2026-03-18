package repository

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/yurisachan16-creator/Memory-system/backend/internal/model"
)

const (
	ListCacheTTL    = 5 * time.Minute
	SearchCacheTTL  = 5 * time.Minute
	SummaryCacheTTL = 10 * time.Minute
	DedupLockTTL    = 10 * time.Second
)

func BuildListCacheKey(filter ListFilter) string {
	params := fmt.Sprintf(
		"category=%s&sort_by=%s&order=%s&page=%d&page_size=%d",
		filter.Category,
		filter.SortBy,
		filter.Order,
		filter.Page,
		filter.PageSize,
	)
	return fmt.Sprintf("memories:list:%s:%s", filter.UserID, hashCacheInput(params))
}

func BuildSearchCacheKey(userID, query string) string {
	return fmt.Sprintf("memories:search:%s:%s", userID, model.HashContent(query))
}

func BuildSummaryCacheKey(userID string) string {
	return fmt.Sprintf("memories:summary:%s", userID)
}

func BuildDedupLockKey(userID, contentHash string) string {
	return fmt.Sprintf("memories:dedup:%s:%s", userID, contentHash)
}

func ListCachePrefix(userID string) string {
	return fmt.Sprintf("memories:list:%s:", userID)
}

func SearchCachePrefix(userID string) string {
	return fmt.Sprintf("memories:search:%s:", userID)
}

func hashCacheInput(value string) string {
	normalized := strings.TrimSpace(value)
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}
