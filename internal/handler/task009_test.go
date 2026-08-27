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

func TestChannelDeleteIsSingleLifecycleTransition(t *testing.T) {
	db := task009DB(t)
	stamp := time.Now().UnixNano()
	owner := &domain.User{Username: fmt.Sprintf("owner-%d", stamp), Email: fmt.Sprintf("owner-%d@example.test", stamp), PasswordHash: "hash", Status: 1}
	if err := db.Create(owner).Error; err != nil {
		t.Fatal(err)
	}
	channel := &domain.Channel{OwnerID: owner.ID, Title: fmt.Sprintf("channel-%d", stamp), Slug: fmt.Sprintf("channel-%d", stamp), Status: domain.ChannelApproved}
	if err := db.Create(channel).Error; err != nil {
		t.Fatal(err)
	}

	h := NewChannelHandler(service.NewChannelService(repository.NewChannelRepository(db)), nil)
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprint(channel.ID)}}
	c.Set(string(middleware.CtxUserID), owner.ID)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/channels/"+fmt.Sprint(channel.ID), nil)
	h.Delete(c)
	if rec.Code != http.StatusOK {
		panic(fmt.Sprintf("channel deletion returned %d, want 200", rec.Code))
	}
}

func task009DB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("BOKE_TEST_DSN")
	if dsn == "" {
		dsn = "host=127.0.0.1 port=55432 user=boke password=boke dbname=boke sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.User{}, &domain.Channel{}); err != nil {
		t.Fatal(err)
	}
	return db
}
