package service

import (
	"fmt"
	"sync"
	"testing"
)

func TestConcurrentCacheInvalidationAdvancesChannelVersion(t *testing.T) {
	rss := NewRSSService(nil, nil, "https://example.test")
	const channelID uint64 = 6006
	var start sync.WaitGroup
	var done sync.WaitGroup
	start.Add(1)
	for i := 0; i < 2; i++ {
		done.Add(1)
		go func() {
			defer done.Done()
			start.Wait()
			for n := 0; n < 100; n++ {
				rss.InvalidateCache(channelID)
			}
		}()
	}
	start.Done()
	done.Wait()
	rss.cacheMu.RLock()
	version := rss.cacheVersion[channelID]
	rss.cacheMu.RUnlock()
	if version == 0 {
		panic(fmt.Sprintf("cache invalidation did not advance channel version for %d", channelID))
	}
}
