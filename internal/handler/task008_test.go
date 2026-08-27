package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"podcast-platform/internal/domain"
	"podcast-platform/internal/repository"
	"podcast-platform/internal/service"
)

var task008SQLiteDriverOnce sync.Once

func TestUnsubscribeKeepsChannelSubscriberCountConsistent(t *testing.T) {
	db := task008DB(t)
	stamp := time.Now().UnixNano()
	owner := &domain.User{Username: fmt.Sprintf("owner-%d", stamp), Email: fmt.Sprintf("owner-%d@example.test", stamp), PasswordHash: "hash", Status: 1}
	if err := db.Create(owner).Error; err != nil {
		t.Fatal(err)
	}
	channel := &domain.Channel{OwnerID: owner.ID, Title: fmt.Sprintf("channel-%d", stamp), Slug: fmt.Sprintf("channel-%d", stamp), Status: domain.ChannelApproved, Subscribers: 1}
	if err := db.Create(channel).Error; err != nil {
		t.Fatal(err)
	}
	sub := &domain.Subscriber{ChannelID: channel.ID, Email: fmt.Sprintf("sub-%d@example.test", stamp), Status: 1, UnsubToken: fmt.Sprintf("token-%d", stamp), SubscribedAt: time.Now()}
	if err := db.Create(sub).Error; err != nil {
		t.Fatal(err)
	}

	svc := service.NewSubscriberService(repository.NewSubscriberRepository(db), repository.NewChannelRepository(db))
	h := NewRSSHandler(nil, repository.NewChannelRepository(db))
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = gin.Params{{Key: "token", Value: sub.UnsubToken}}
	c.Set("subscriber_svc", svc)
	c.Request = httptest.NewRequest(http.MethodGet, "/unsubscribe/"+sub.UnsubToken, nil)
	h.Unsubscribe(c)
	if rec.Code != http.StatusOK {
		t.Fatalf("unsubscribe returned %d: %s", rec.Code, rec.Body.String())
	}
	var got domain.Channel
	if err := db.First(&got, channel.ID).Error; err != nil {
		t.Fatal(err)
	}
	if got.Subscribers != 0 {
		panic(fmt.Sprintf("channel subscriber count is %d, want 0", got.Subscribers))
	}
}

func task008DB(t *testing.T) *gorm.DB {
	t.Helper()
	// Keep the contract test self-contained so Red and Green run identically.
	task008SQLiteDriverOnce.Do(func() {
		sql.Register("task008_sqlite", &sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				return conn.RegisterFunc("GREATEST", func(left, right int64) int64 {
					if left > right {
						return left
					}
					return right
				}, true)
			},
		})
	})
	sqlDB, err := sql.Open("task008_sqlite", "file:task008?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.User{}, &domain.Channel{}, &domain.Subscriber{}); err != nil {
		t.Fatal(err)
	}
	return db
}
