package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"podcast-platform/internal/domain"
	"podcast-platform/internal/middleware"
	"podcast-platform/internal/repository"
	"podcast-platform/internal/service"
)

func TestCommentApprovalKeepsEpisodeCountConsistent(t *testing.T) {
	db := task007DB(t)
	stamp := time.Now().UnixNano()
	owner := &domain.User{Username: fmt.Sprintf("owner-%d", stamp), Email: fmt.Sprintf("owner-%d@example.test", stamp), PasswordHash: "hash", Status: 1}
	if err := db.Create(owner).Error; err != nil {
		t.Fatal(err)
	}
	channel := &domain.Channel{OwnerID: owner.ID, Title: fmt.Sprintf("channel-%d", stamp), Slug: fmt.Sprintf("channel-%d", stamp), Status: domain.ChannelApproved}
	if err := db.Create(channel).Error; err != nil {
		t.Fatal(err)
	}
	episode := &domain.Episode{ChannelID: channel.ID, Title: "episode", Slug: fmt.Sprintf("episode-%d", time.Now().UnixNano()), AudioFileURL: "audio.mp3", Status: domain.EpisodePublished}
	if err := db.Create(episode).Error; err != nil {
		t.Fatal(err)
	}
	comment := &domain.Comment{ChannelID: channel.ID, EpisodeID: episode.ID, UserID: owner.ID, Type: domain.CommentTypeEpisode, Content: "hello", Status: domain.CommentPending}
	if err := db.Create(comment).Error; err != nil {
		t.Fatal(err)
	}

	comments := repository.NewCommentRepository(db)
	episodes := repository.NewEpisodeRepository(db)
	channels := repository.NewChannelRepository(db)
	h := NewStatsHandler(nil, service.NewCommentService(comments, episodes, channels))
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = gin.Params{{Key: "comment_id", Value: fmt.Sprint(comment.ID)}}
	c.Set(string(middleware.CtxUserID), uint64(1))
	c.Set(string(middleware.CtxRole), domain.RoleAdmin)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/comments/approve", nil)
	h.ApproveComment(c)
	if rec.Code != http.StatusOK {
		t.Fatalf("approval returned %d: %s", rec.Code, rec.Body.String())
	}

	var got domain.Episode
	if err := db.First(&got, episode.ID).Error; err != nil {
		t.Fatal(err)
	}
	if got.CommentCount != 1 {
		panic(fmt.Sprintf("approved comment count is %d, want 1", got.CommentCount))
	}
}

func task007DB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("BOKE_TEST_DSN")
	if dsn == "" {
		dsn = "host=127.0.0.1 port=55432 user=boke password=boke dbname=boke sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.User{}, &domain.Channel{}, &domain.Episode{}, &domain.Comment{}); err != nil {
		t.Fatal(err)
	}
	return db
}
