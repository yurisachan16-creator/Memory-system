package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yurisachan16-creator/Memory-system/backend/internal/response"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		response.Error(c, http.StatusInternalServerError, 1000, "internal server error")
		c.Abort()
	})
}

