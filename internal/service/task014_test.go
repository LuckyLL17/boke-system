package service

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"podcast-platform/internal/domain"
	"podcast-platform/internal/repository"
)

// AggregateHandler.Run -> StatsService.UpdateDailyCache -> PlaybackRepository.SaveStatsCache
func TestHistoricalCacheUsesRequestedDate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:task014?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.StatsCache{}); err != nil {
		t.Fatal(err)
	}
	wanted := time.Now().AddDate(0, 0, -1)
	svc := &StatsService{playbackRepo: repository.NewPlaybackRepository(db)}
	if err := svc.UpdateDailyCache(8, wanted, DailySummaryRow{Date: wanted, TotalPlaybacks: 4, UniqueUsers: 2}); err != nil {
		t.Fatal(err)
	}
	var cache domain.StatsCache
	if err := db.First(&cache).Error; err != nil {
		t.Fatal(err)
	}
	if !cache.Date.Truncate(24 * time.Hour).Equal(wanted.Truncate(24 * time.Hour)) {
		t.Fatalf("historical cache used the wrong date: got=%s want=%s", cache.Date, wanted)
	}
}
