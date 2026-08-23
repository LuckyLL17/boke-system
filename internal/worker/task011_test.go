package worker

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"podcast-platform/internal/domain"
	"podcast-platform/internal/repository"
	"podcast-platform/internal/service"
	"podcast-platform/pkg/logger"
)

// PublishHandler.Run -> EpisodeService.ProcessScheduledEpisodes -> EpisodeRepository.UpdateStatus
func TestScheduledPublishClearsScheduleState(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:task011?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.Episode{}); err != nil {
		t.Fatal(err)
	}
	scheduled := time.Now().Add(-time.Minute)
	ep := &domain.Episode{ChannelID: 7, Title: "夜间节目", AudioFileURL: "/audio/night.mp3", Status: domain.EpisodeScheduled, ScheduledAt: &scheduled}
	if err := db.Create(ep).Error; err != nil {
		t.Fatal(err)
	}
	repo := repository.NewEpisodeRepository(db)
	h := NewPublishHandler(repo, service.NewRSSService(nil, nil, "https://example.test"), logger.New(true))
	if err := h.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	var got domain.Episode
	if err := db.First(&got, ep.ID).Error; err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.EpisodePublished || got.PublishedAt == nil || got.ScheduledAt != nil {
		t.Fatalf("published episode state was inconsistent: status=%q scheduled_at=%v", got.Status, got.ScheduledAt)
	}
}
