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

func TestChannelApprovalPropagatesStorageFailure(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:task023?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	h := NewChannelHandler(
		service.NewChannelService(repository.NewChannelRepository(db)),
		nil,
	)
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = gin.Params{{Key: "id", Value: "11"}}
	h.Approve(c)
	if rec.Code == 200 {
		t.Fatalf("approval storage failure was reported as success")
	}
}
