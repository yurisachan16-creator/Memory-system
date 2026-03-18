package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRateLimitRejectsRequestsBeyondBurst(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ipLimiters = sync.Map{}

	router := gin.New()
	router.Use(RateLimit())
	router.GET("/limited", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	var limitedResponse *httptest.ResponseRecorder
	for i := 0; i < rateLimitBurst+20; i++ {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/limited", nil)
		request.RemoteAddr = "198.51.100.10:12345"
		router.ServeHTTP(response, request)
		if response.Code == http.StatusTooManyRequests {
			limitedResponse = response
			break
		}
	}

	if limitedResponse == nil {
		t.Fatalf("expected at least one request to hit rate limit")
	}

	var payload struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(limitedResponse.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode rate limit payload: %v", err)
	}
	if payload.Code != http.StatusTooManyRequests || payload.Message != "rate limit exceeded" {
		t.Fatalf("unexpected rate limit payload: %+v", payload)
	}
}
