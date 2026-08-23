package repository

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"podcast-platform/internal/domain"
)

// Update -> ProcessScheduledEpisodes -> PublishHandler.Run -> EpisodeRepository.UpdateStatus.
func TestPublishStatusPreservesScheduledTime(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:task029?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.Episode{}); err != nil {
		t.Fatal(err)
	}
	scheduled := time.Now().Add(-time.Hour).Truncate(time.Second)
	episode := &domain.Episode{
		ChannelID: 1, Title: "Scheduled", Slug: "scheduled-029",
		AudioFileURL: "/audio.mp3", Status: domain.EpisodeScheduled,
		ScheduledAt: &scheduled,
	}
	if err := db.Create(episode).Error; err != nil {
		t.Fatal(err)
	}
	repo := NewEpisodeRepository(db)
	if err := repo.UpdateStatus(episode.ID, string(domain.EpisodePublished)); err != nil {
		t.Fatal(err)
	}
	var saved domain.Episode
	if err := db.First(&saved, episode.ID).Error; err != nil {
		t.Fatal(err)
	}
	if saved.ScheduledAt == nil || !saved.ScheduledAt.Equal(scheduled) {
		t.Fatalf("publishing changed the existing schedule time")
	}
}
