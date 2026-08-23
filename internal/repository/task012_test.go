package repository

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"podcast-platform/internal/domain"
)

// EpisodeHandler.Create -> EpisodeService.Create -> ChapterRepository.BatchCreate
func TestBatchCreatePersistsEveryChapter(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:task012?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.Chapter{}); err != nil {
		t.Fatal(err)
	}
	chapters := []domain.Chapter{{EpisodeID: 9, Title: "一", StartTime: 0}, {EpisodeID: 9, Title: "二", StartTime: 60}, {EpisodeID: 9, Title: "三", StartTime: 120}}
	if err := NewChapterRepository(db).BatchCreate(chapters); err != nil {
		t.Fatal(err)
	}
	var got []domain.Chapter
	if err := db.Where("episode_id = ?", 9).Find(&got).Error; err != nil {
		t.Fatal(err)
	}
	if len(got) != len(chapters) {
		t.Fatalf("chapter persistence was incomplete: got %d want %d", len(got), len(chapters))
	}
}
