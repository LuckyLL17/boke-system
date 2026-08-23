package service

import (
	"encoding/xml"
	"sync"
	"time"

	"podcast-platform/config"
	"podcast-platform/internal/domain"
	"podcast-platform/internal/repository"
	appErr "podcast-platform/pkg/errors"
	pkgRss "podcast-platform/pkg/rss"
)

type RSSService struct {
	channelRepo  *repository.ChannelRepository
	episodeRepo  *repository.EpisodeRepository
	generator    *pkgRss.Generator
	cache        map[uint64]*RSSCache
	cacheVersion map[uint64]uint64
	cacheMu      sync.RWMutex
	baseSiteURL  string
}

type RSSCache struct {
	Content   string
	Generated time.Time
	ExpiresAt time.Time
}

func NewRSSService(channelRepo *repository.ChannelRepository, episodeRepo *repository.EpisodeRepository, baseURL string) *RSSService {
	return &RSSService{
		channelRepo:  channelRepo,
		episodeRepo:  episodeRepo,
		generator:    pkgRss.NewGenerator(baseURL),
		cache:        make(map[uint64]*RSSCache),
		cacheVersion: make(map[uint64]uint64),
		baseSiteURL:  baseURL,
	}
}

func (s *RSSService) GenerateFeed(channelID uint64, useCache bool) (string, error) {
	if useCache {
		s.cacheMu.RLock()
		if cached, ok := s.cache[channelID]; ok && cached.ExpiresAt.After(time.Now()) {
			content := cached.Content
			s.cacheMu.RUnlock()
			return content, nil
		}
		s.cacheMu.RUnlock()
	}
	// Rebuild under the write lock so concurrent rebuilds serialize: only one
	// goroutine builds at a time, the rest wait and then observe the result via
	// the double-checked cache read below. This closes the read-build-write
	// window where a stale rebuild could overwrite a fresher writer.
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	if useCache {
		if cached, ok := s.cache[channelID]; ok && cached.ExpiresAt.After(time.Now()) {
			return cached.Content, nil
		}
	}
	// Snapshot the cache version; if InvalidateCache bumped it while we rebuilt,
	// another writer (e.g. a scheduled-publish) owns the entry and we must not
	// clobber it with our possibly-stale snapshot.
	version := s.cacheVersion[channelID]
	feed, err := s.buildFeed(channelID)
	if err != nil {
		return "", err
	}
	if s.cacheVersion[channelID] != version {
		return feed, nil
	}
	ttl := time.Duration(config.AppConfig.Cache.RSSTTL) * time.Second
	s.cache[channelID] = &RSSCache{
		Content:   feed,
		Generated: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}
	return feed, nil
}

// buildFeed reads the channel and its published episodes from the database and
// renders the RSS XML. It performs no caching and must be called while the
// caller holds the cache write lock (or with no caching expectations).
func (s *RSSService) buildFeed(channelID uint64) (string, error) {
	ch, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return "", appErr.ErrChannelNotFound
	}
	if ch.Status != domain.ChannelApproved {
		return "", appErr.ErrForbidden
	}
	episodes, err := s.episodeRepo.ListPublishedByChannel(channelID, 500)
	if err != nil {
		return "", appErr.Wrap(err, 500, "load episodes")
	}
	return s.generator.GenerateFeed(ch, episodes, s.baseSiteURL), nil
}

// InvalidateCache drops the cached feed for a channel and bumps its version.
// The version bump is what makes the CAS check in GenerateFeed meaningful: any
// in-flight rebuild that snapshotted the previous version will, on write, see
// the version changed and discard its result instead of clobbering a fresher
// entry. Call this after any state change that should be reflected in the feed
// (episode published, channel metadata edited, etc.).
func (s *RSSService) InvalidateCache(channelID uint64) {
	s.cacheMu.Lock()
	delete(s.cache, channelID)
	s.cacheVersion[channelID]++
	s.cacheMu.Unlock()
}

func (s *RSSService) RefreshAllCaches() (int, error) {
	ids, err := s.channelRepo.GetAllIDs(1000)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, id := range ids {
		if _, err := s.GenerateFeed(id, false); err == nil {
			count++
		}
	}
	return count, nil
}

func (s *RSSService) ValidateFeed(channelID uint64) ([]string, error) {
	ch, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return nil, appErr.ErrChannelNotFound
	}
	episodes, err := s.episodeRepo.ListPublishedByChannel(channelID, 100)
	if err != nil {
		return nil, err
	}
	issues := pkgRss.ValidateFeed(ch, episodes)
	return issues, nil
}

func (s *RSSService) GetSubscribeLinks(channelID uint64, slug string) (map[string]string, error) {
	ch, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return nil, appErr.ErrChannelNotFound
	}
	_ = ch
	base := s.baseSiteURL
	if slug == "" {
		slug = "channel-" + uint64Str(channelID)
	}
	rssURL := base + "/api/v1/channels/" + uint64Str(channelID) + "/rss"
	links := map[string]string{
		"rss":        rssURL,
		"apple":      "https://podcasts.apple.com/subscribe?url=" + rssURL,
		"spotify":    "https://open.spotify.com/search/" + rssURL,
		"google":     "https://podcasts.google.com/?feed=" + rssURL,
		"xiaoyuzhou": "https://www.xiaoyuzhoufm.com/podcast?url=" + rssURL,
		"slug_rss":   base + "/rss/" + slug + ".xml",
	}
	return links, nil
}

func uint64Str(n uint64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}

func (s *RSSService) IsValidXML(content string) bool {
	return xml.Unmarshal([]byte(content), new(interface{})) == nil
}
