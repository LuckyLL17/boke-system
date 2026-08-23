package repository

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"podcast-platform/internal/domain"
)

// EpisodeHandler.AddComment -> CommentService.Create -> CommentRepository.ListByEpisode
func TestEpisodeCommentsUseEpisodeAndNotChannelID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:task016?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.Comment{}); err != nil {
		t.Fatal(err)
	}
	comment := &domain.Comment{EpisodeID: 7, ChannelID: 99, UserID: 1, Type: domain.CommentTypeEpisode, Content: "很好听", Status: domain.CommentApproved}
	if err := db.Create(comment).Error; err != nil {
		t.Fatal(err)
	}
	status := domain.CommentApproved
	comments, total, err := NewCommentRepository(db).ListByEpisode(7, 1, 20, &status)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(comments) != 1 {
		t.Fatalf("episode comment was hidden by wrong ownership filter: total=%d rows=%d", total, len(comments))
	}
}
