package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yurisachan16-creator/Memory-system/backend/internal/repository"
	"github.com/yurisachan16-creator/Memory-system/backend/internal/service"
)

func newTestServer(now time.Time) http.Handler {
	gin.SetMode(gin.TestMode)

	svc := service.NewMemoryService(repository.NewInMemoryMemoryRepository(), repository.NewInMemoryCache())
	svc.SetNowFunc(func() time.Time { return now })
	engine := gin.New()
	api := engine.Group("/api/v1")
	RegisterMemoryRoutes(api, svc)
	return engine
}

func TestMemoryCRUDSearchAndSummaryFlow(t *testing.T) {
	server := newTestServer(time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC))

	createBody := map[string]interface{}{
		"user_id":    "u1",
		"content":    "User likes espresso drinks",
		"category":   "preference",
		"source":     "manual",
		"importance": 4,
	}
	createResp := performJSON(t, server, http.MethodPost, "/api/v1/memories", createBody)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create, got %d", createResp.Code)
	}

	listResp := performJSON(t, server, http.MethodGet, "/api/v1/memories?user_id=u1&page=1&page_size=10", nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected 200 on list, got %d", listResp.Code)
	}

	var listPayload struct {
		Code int `json:"code"`
		Data struct {
			Items []struct {
				ID int64 `json:"id"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(listResp.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("decode list payload: %v", err)
	}
	if len(listPayload.Data.Items) != 1 {
		t.Fatalf("expected one list item, got %d", len(listPayload.Data.Items))
	}

	updateBody := map[string]interface{}{
		"user_id":    "u1",
		"content":    "User likes espresso and cappuccino",
		"category":   "preference",
		"importance": 5,
	}
	updateResp := performJSON(t, server, http.MethodPut, "/api/v1/memories/"+strconv.FormatInt(listPayload.Data.Items[0].ID, 10), updateBody)
	if updateResp.Code != http.StatusOK {
		t.Fatalf("expected 200 on update, got %d", updateResp.Code)
	}

	searchResp := performJSON(t, server, http.MethodGet, "/api/v1/memories/search?user_id=u1&query=espresso", nil)
	if searchResp.Code != http.StatusOK {
		t.Fatalf("expected 200 on search, got %d", searchResp.Code)
	}

	summaryResp := performJSON(t, server, http.MethodGet, "/api/v1/memories/summary?user_id=u1", nil)
	if summaryResp.Code != http.StatusOK {
		t.Fatalf("expected 200 on summary, got %d", summaryResp.Code)
	}

	deleteResp := performJSON(t, server, http.MethodDelete, "/api/v1/memories/"+strconv.FormatInt(listPayload.Data.Items[0].ID, 10)+"?user_id=u1", nil)
	if deleteResp.Code != http.StatusOK {
		t.Fatalf("expected 200 on delete, got %d", deleteResp.Code)
	}
}

func TestUpdateRejectsWrongOwner(t *testing.T) {
	server := newTestServer(time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC))

	createBody := map[string]interface{}{
		"user_id":    "owner",
		"content":    "Owner memory",
		"category":   "context",
		"source":     "manual",
		"importance": 3,
	}
	if resp := performJSON(t, server, http.MethodPost, "/api/v1/memories", createBody); resp.Code != http.StatusCreated {
		t.Fatalf("expected create ok, got %d", resp.Code)
	}

	listResp := performJSON(t, server, http.MethodGet, "/api/v1/memories?user_id=owner&page=1&page_size=10", nil)
	var listPayload struct {
		Data struct {
			Items []struct {
				ID int64 `json:"id"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(listResp.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("decode list payload: %v", err)
	}

	updateBody := map[string]interface{}{
		"user_id":    "intruder",
		"content":    "Intruder update",
		"category":   "context",
		"importance": 5,
	}
	resp := performJSON(t, server, http.MethodPut, "/api/v1/memories/"+strconv.FormatInt(listPayload.Data.Items[0].ID, 10), updateBody)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for wrong owner, got %d", resp.Code)
	}
}

func performJSON(t *testing.T, handler http.Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}
