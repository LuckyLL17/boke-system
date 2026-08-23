package rss

import (
	"strings"
	"testing"

	"podcast-platform/internal/domain"
)

// RSSHandler.Feed -> RSSService.GenerateFeed -> Generator.GenerateFeed
func TestRSSRefreshUsesCurrentSiteURL(t *testing.T) {
	g := NewGenerator("https://internal.example")
	channel := &domain.Channel{ID: 4, Title: "频道", Description: "描述", Language: "zh-CN", Author: "作者"}
	episodes := []domain.Episode{{ID: 9, Title: "节目", AudioFileURL: "/audio/episode.mp3"}}
	feed := g.GenerateFeed(channel, episodes, "https://public.example")
	if !strings.Contains(feed, "https://public.example/episode/9") {
		t.Fatalf("RSS feed used a stale site URL: %s", feed)
	}
}
