package handler

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"podcast-platform/internal/domain"
	"podcast-platform/internal/repository"
	"podcast-platform/internal/service"
)

func TestCustomDomainReachesSubscribeLinks(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:task028?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.User{}, &domain.Category{}, &domain.Channel{}); err != nil {
		t.Fatal(err)
	}
	owner := &domain.User{Username: "owner028", Email: "owner028@example.com", PasswordHash: "hash", Role: domain.RoleUser, Status: 1}
	if err := db.Create(owner).Error; err != nil {
		t.Fatal(err)
	}
	channel := &domain.Channel{OwnerID: owner.ID, Title: "Channel", Slug: "channel-028"}
	if err := db.Create(channel).Error; err != nil {
		t.Fatal(err)
	}
	channelRepo := repository.NewChannelRepository(db)
	channelSvc := service.NewChannelService(channelRepo)
	h := NewChannelHandler(channelSvc, nil)
	body, _ := json.Marshal(map[string]string{"custom_domain": "https://listen.example.test"})
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", owner.ID)
	c.Request = httptest.NewRequest("PUT", "/api/v1/channels/1", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Update(c)
	if rec.Code != 200 {
		t.Fatalf("channel update returned %d: %s", rec.Code, rec.Body.String())
	}

	rss := service.NewRSSService(channelRepo, nil, "https://platform.example.test")
	links, err := rss.GetSubscribeLinks(channel.ID, channel.Slug)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(links["rss"], "https://listen.example.test/") ||
		!strings.HasPrefix(links["slug_rss"], "https://listen.example.test/") {
		t.Fatalf("subscribe link ignored the configured custom domain")
	}
}
