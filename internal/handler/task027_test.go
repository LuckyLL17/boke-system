package handler

import (
	"encoding/json"
	"math"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"podcast-platform/internal/domain"
	"podcast-platform/internal/repository"
	"podcast-platform/internal/service"
)

func TestEpisodeCompletionReportsPlaybackStatistics(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:task027?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.Playback{}); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	playbacks := []domain.Playback{
		{EpisodeID: 7, ChannelID: 3, IPAddress: "10.0.0.1", StartAt: now, Progress: 0.8, Completed: true},
		{EpisodeID: 7, ChannelID: 3, IPAddress: "10.0.0.2", StartAt: now, Progress: 0.2, Completed: false},
	}
	if err := db.Create(&playbacks).Error; err != nil {
		t.Fatal(err)
	}
	stats := service.NewStatsService(
		repository.NewPlaybackRepository(db),
		nil,
		nil,
		nil,
	)
	h := NewStatsHandler(stats, nil)
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	h.EpisodeCompletion(c)
	if rec.Code != 200 {
		t.Fatalf("completion endpoint returned %d: %s", rec.Code, rec.Body.String())
	}
	var response struct {
		Data struct {
			AvgProgress    float64 `json:"avg_progress"`
			CompletionRate float64 `json:"completion_rate"`
			TotalPlays     int64   `json:"total_plays"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if math.Abs(response.Data.AvgProgress-0.5) > 0.0001 ||
		math.Abs(response.Data.CompletionRate-0.5) > 0.0001 ||
		response.Data.TotalPlays != 2 {
		t.Fatalf("completion statistics were inconsistent")
	}
}
