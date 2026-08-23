package handler

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"podcast-platform/internal/domain"
	"podcast-platform/internal/repository"
	"podcast-platform/internal/service"
)

func TestEpisodeUpdateClearsSubmittedChapters(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:task024?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.User{}, &domain.Category{}, &domain.Channel{}, &domain.Episode{}, &domain.Chapter{}); err != nil {
		t.Fatal(err)
	}
	owner := &domain.User{Username: "owner024", Email: "owner024@example.com", PasswordHash: "hash", Role: domain.RoleUser, Status: 1}
	if err := db.Create(owner).Error; err != nil {
		t.Fatal(err)
	}
	channel := &domain.Channel{OwnerID: owner.ID, Title: "Channel", Slug: "channel-024"}
	if err := db.Create(channel).Error; err != nil {
		t.Fatal(err)
	}
	episode := &domain.Episode{ChannelID: channel.ID, Title: "Episode", Slug: "episode-024", AudioFileURL: "/audio.mp3", Status: domain.EpisodeDraft}
	if err := db.Create(episode).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&domain.Chapter{EpisodeID: episode.ID, Title: "Old", StartTime: 1}).Error; err != nil {
		t.Fatal(err)
	}

	svc := service.NewEpisodeService(
		repository.NewEpisodeRepository(db),
		repository.NewChapterRepository(db),
		repository.NewChannelRepository(db),
		nil,
	)
	h := NewEpisodeHandler(svc, nil, nil, nil)
	body, _ := json.Marshal(map[string]string{"chapters_json": "[]"})
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", owner.ID)
	c.Request = httptest.NewRequest("PUT", "/api/v1/episodes/1", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Update(c)
	if rec.Code != 200 {
		t.Fatalf("update returned %d: %s", rec.Code, rec.Body.String())
	}
	var count int64
	if err := db.Model(&domain.Chapter{}).Where("episode_id = ?", episode.ID).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("empty chapter submission left old chapters")
	}
}
