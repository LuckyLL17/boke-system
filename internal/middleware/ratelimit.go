package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int
	window   time.Duration
}

type visitor struct {
	count     int
	lastCheck time.Time
}

func NewRateLimiter(ratePerMin int) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     ratePerMin,
		window:   time.Minute,
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	go rl.cleanup()
	return func(c *gin.Context) {
		key := c.ClientIP()
		rl.mu.Lock()
		v, ok := rl.visitors[key]
		if !ok {
			v = &visitor{count: 1, lastCheck: time.Now()}
			rl.visitors[key] = v
			rl.mu.Unlock()
			c.Next()
			return
		}
		now := time.Now()
		elapsed := now.Sub(v.lastCheck)
		if elapsed > rl.window {
			v.count = 1
			v.lastCheck = now
			rl.mu.Unlock()
			c.Next()
			return
		}
		v.count++
		if v.count > rl.rate {
			rl.mu.Unlock()
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "rate limit exceeded",
				"error":   true,
			})
			return
		}
		rl.mu.Unlock()
		c.Next()
	}
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for k, v := range rl.visitors {
			if now.Sub(v.lastCheck) > rl.window*2 {
				delete(rl.visitors, k)
			}
		}
		rl.mu.Unlock()
	}
}

func PerUserRateLimit(ratePerMin int) gin.HandlerFunc {
	rl := NewRateLimiter(ratePerMin)
	return rl.Middleware()
}

func GlobalRateLimit() gin.HandlerFunc {
	return PerUserRateLimit(120)
}
