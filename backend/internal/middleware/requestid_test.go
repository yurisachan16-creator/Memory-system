package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestRequestIDSetsHeaderContextAndLoggerOutput(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logBuffer bytes.Buffer
	originalWriter := log.Writer()
	log.SetOutput(&logBuffer)
	defer log.SetOutput(originalWriter)

	router := gin.New()
	router.Use(RequestID(), Logger())
	router.GET("/ping", func(c *gin.Context) {
		requestID, ok := c.Get("request_id")
		if !ok {
			t.Fatalf("expected request_id in gin context")
		}
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	requestID := response.Header().Get("X-Request-Id")
	if requestID == "" {
		t.Fatalf("expected X-Request-Id header")
	}
	if _, err := uuid.Parse(requestID); err != nil {
		t.Fatalf("expected valid uuid, got %q", requestID)
	}
	if !strings.Contains(logBuffer.String(), "request_id="+requestID) {
		t.Fatalf("expected logger output to include request_id, got %q", logBuffer.String())
	}
}
