package middleware

import (
	"net/http/httptest"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// GlobalRateLimit -> RateLimiter.Middleware -> cleanup goroutine lifecycle
func TestRateLimiterStartsOnlyOneCleanupLoop(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := NewRateLimiter(100)
	h := rl.Middleware()
	before := runtime.NumGoroutine()
	var wg sync.WaitGroup
	wg.Add(8)
	for i := 0; i < 8; i++ {
		go func() {
			defer wg.Done()
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest("GET", "/", nil)
			h(c)
		}()
	}
	wg.Wait()
	time.Sleep(20 * time.Millisecond)
	after := runtime.NumGoroutine()
	if after-before > 3 {
		t.Fatalf("rate limiter started too many cleanup goroutines: before=%d after=%d", before, after)
	}
}
