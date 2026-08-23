package service

import (
	"sync"
	"testing"
	"time"
)

// RefreshCache -> GenerateFeed -> InvalidateCache -> storeFeed
func TestStaleFeedVersionCannotReplaceNewerCache(t *testing.T) {
	s := &RSSService{
		cache:        make(map[uint64]*RSSCache),
		cacheVersion: make(map[uint64]uint64),
	}
	s.cacheVersion[7] = 2
	start := make(chan struct{})
	releaseOlder := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		<-releaseOlder
		s.storeFeed(7, "older-feed", 1, time.Hour)
	}()
	newDone := make(chan struct{})
	go func() {
		defer wg.Done()
		<-start
		s.storeFeed(7, "newer-feed", 2, time.Hour)
		close(newDone)
	}()
	close(start)
	<-newDone
	close(releaseOlder)
	wg.Wait()
	s.cacheMu.RLock()
	cached, stored := s.cache[7]
	s.cacheMu.RUnlock()
	if !stored || cached.Content != "newer-feed" {
		panic("stale feed version replaced newer cache")
	}
}
