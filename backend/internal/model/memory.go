package model

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

type Category string

const (
	CategoryPreference Category = "preference"
	CategoryIdentity   Category = "identity"
	CategoryGoal       Category = "goal"
	CategoryContext    Category = "context"
)

func (c Category) Valid() bool {
	switch c {
	case CategoryPreference, CategoryIdentity, CategoryGoal, CategoryContext:
		return true
	default:
		return false
	}
}

type Source string

const (
	SourceChat   Source = "chat"
	SourceManual Source = "manual"
	SourceSystem Source = "system"
)

func (s Source) Valid() bool {
	switch s {
	case SourceChat, SourceManual, SourceSystem:
		return true
	default:
		return false
	}
}

type Memory struct {
	ID          int64     `json:"id"`
	UserID      string    `json:"user_id"`
	Content     string    `json:"content"`
	Category    Category  `json:"category"`
	Source      Source    `json:"source"`
	Importance  int       `json:"importance"`
	ContentHash string    `json:"content_hash"`
	IsDeleted   bool      `json:"is_deleted"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (m Memory) Clone() Memory {
	return m
}

type CreateMemoryRequest struct {
	UserID     string   `json:"user_id"`
	Content    string   `json:"content"`
	Category   Category `json:"category"`
	Source     Source   `json:"source"`
	Importance int      `json:"importance"`
}

type UpdateMemoryRequest struct {
	UserID     string   `json:"user_id"`
	Content    string   `json:"content"`
	Category   Category `json:"category"`
	Importance int      `json:"importance"`
}

type ListMemoriesQuery struct {
	UserID   string
	Category Category
	SortBy   string
	Order    string
	Page     int
	PageSize int
}

type ListMemoriesResult struct {
	Items    []Memory `json:"items"`
	Total    int      `json:"total"`
	Page     int      `json:"page"`
	PageSize int      `json:"page_size"`
}

type SearchMemoryResult struct {
	Memory         Memory   `json:"memory"`
	RelevanceScore float64  `json:"relevance_score"`
	RecencyScore   float64  `json:"recency_score"`
	FinalScore     float64  `json:"final_score"`
	MatchedTerms   []string `json:"matched_terms"`
}

type SummaryResult struct {
	Preferences []Memory `json:"preferences"`
	Goals       []Memory `json:"goals"`
	Background  []Memory `json:"background"`
	Recent      []Memory `json:"recent"`
}

func NormalizeContent(content string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(content)), " ")
}

func HashContent(content string) string {
	normalized := strings.ToLower(NormalizeContent(content))
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}
