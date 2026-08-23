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

func TestPartialChannelUpdatePreservesExistingFields(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:task025?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.User{}, &domain.Category{}, &domain.Channel{}); err != nil {
		t.Fatal(err)
	}
	owner := &domain.User{Username: "owner025", Email: "owner025@example.com", PasswordHash: "hash", Role: domain.RoleUser, Status: 1}
	if err := db.Create(owner).Error; err != nil {
		t.Fatal(err)
	}
	channel := &domain.Channel{
		OwnerID: owner.ID, Title: "Before", Slug: "channel-025",
		Description: "Keep this description", CoverImageURL: "https://img.example/cover.png",
	}
	if err := db.Create(channel).Error; err != nil {
		t.Fatal(err)
	}

	h := NewChannelHandler(
		service.NewChannelService(repository.NewChannelRepository(db)),
		nil,
	)
	body, _ := json.Marshal(map[string]string{"title": "After"})
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", owner.ID)
	c.Request = httptest.NewRequest("PUT", "/api/v1/channels/1", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Update(c)
	if rec.Code != 200 {
		t.Fatalf("update returned %d: %s", rec.Code, rec.Body.String())
	}
	var saved domain.Channel
	if err := db.First(&saved, channel.ID).Error; err != nil {
		t.Fatal(err)
	}
	if saved.Description != channel.Description || saved.CoverImageURL != channel.CoverImageURL {
		t.Fatalf("partial update cleared an untouched field")
	}
}
