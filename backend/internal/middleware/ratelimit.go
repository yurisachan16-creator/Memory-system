package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const (
	rateLimitPerSecond = 30
	rateLimitBurst     = 50
)

var ipLimiters sync.Map

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := getIPLimiter(ip)
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}

func getIPLimiter(ip string) *rate.Limiter {
	limiter, ok := ipLimiters.Load(ip)
	if ok {
		return limiter.(*rate.Limiter)
	}

	newLimiter := rate.NewLimiter(rate.Limit(rateLimitPerSecond), rateLimitBurst)
	actual, _ := ipLimiters.LoadOrStore(ip, newLimiter)
	return actual.(*rate.Limiter)
}
