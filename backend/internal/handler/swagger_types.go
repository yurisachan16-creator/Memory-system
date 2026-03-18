package handler

import "github.com/yurisachan16-creator/Memory-system/backend/internal/model"

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type createMemoryResponse struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    model.Memory `json:"data"`
}

type listMemoriesResponse struct {
	Code    int                      `json:"code"`
	Message string                   `json:"message"`
	Data    model.ListMemoriesResult `json:"data"`
}

type deleteMemoryResponse struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    deleteMemoryData `json:"data"`
}

type deleteMemoryData struct {
	ID int64 `json:"id"`
}

type searchMemoriesResponse struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    searchMemoriesData `json:"data"`
}

type searchMemoriesData struct {
	Items  []model.SearchMemoryResult `json:"items"`
	Count  int                        `json:"count"`
	Cached bool                       `json:"cached"`
}

type summaryResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    summaryDataWrap `json:"data"`
}

type summaryDataWrap struct {
	Summary model.SummaryResult `json:"summary"`
	Cached  bool                `json:"cached"`
}

type healthResponse struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    healthResponseData `json:"data"`
}

type healthResponseData struct {
	Service  string `json:"service"`
	Database string `json:"database"`
	Redis    string `json:"redis"`
}
