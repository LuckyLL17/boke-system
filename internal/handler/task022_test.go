package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"podcast-platform/internal/repository"
	"podcast-platform/internal/service"
)

func TestEpisodeDeleteReportsServiceFailure(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:task022?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	episodeSvc := service.NewEpisodeService(
		repository.NewEpisodeRepository(db),
		repository.NewChapterRepository(db),
		repository.NewChannelRepository(db),
		nil,
	)
	h := NewEpisodeHandler(episodeSvc, nil, nil, nil)
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = gin.Params{{Key: "id", Value: "9"}}
	c.Set("user_id", uint64(1))
	h.Delete(c)
	if rec.Code == 200 {
		t.Fatalf("delete failure was reported as success")
	}
}
