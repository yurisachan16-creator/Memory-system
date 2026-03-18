package response

import "github.com/gin-gonic/gin"

type Envelope struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, message string, data interface{}) {
	SuccessWithStatus(c, 200, message, data)
}

func SuccessWithStatus(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, Envelope{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, status int, code int, message string) {
	c.JSON(status, Envelope{
		Code:    code,
		Message: message,
	})
}
