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
	stop     chan struct{}
	stopped  chan struct{}
}

type visitor struct {
	count     int
	lastCheck time.Time
}

func NewRateLimiter(ratePerMin int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     ratePerMin,
		window:   time.Minute,
		stop:     make(chan struct{}),
		stopped:  make(chan struct{}),
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		rl.mu.Lock()
		v, ok := rl.visitors[key]
		if !ok {
			rl.visitors[key] = &visitor{count: 1, lastCheck: time.Now()}
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

func (rl *RateLimiter) cleanupLoop() {
	defer close(rl.stopped)
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-rl.stop:
			return
		case <-ticker.C:
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
}

func (rl *RateLimiter) Close() {
	select {
	case <-rl.stop:
		return
	default:
		close(rl.stop)
	}
	<-rl.stopped
}

func PerUserRateLimit(ratePerMin int) (*RateLimiter, gin.HandlerFunc) {
	rl := NewRateLimiter(ratePerMin)
	return rl, rl.Middleware()
}

func GlobalRateLimit() (*RateLimiter, gin.HandlerFunc) {
	return PerUserRateLimit(120)
}
