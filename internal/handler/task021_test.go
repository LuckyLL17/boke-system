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

// POST /api/v1/auth/register -> Register -> ExistsByEmail -> Create.
func TestRegistrationStopsWhenEmailCheckFails(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:task021?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.User{}); err != nil {
		t.Fatal(err)
	}
	queryCount := 0
	if err := db.Callback().Query().Before("gorm:query").Register("task021_fail_email_check", func(tx *gorm.DB) {
		queryCount++
		if queryCount == 2 {
			tx.Error = gorm.ErrInvalidData
		}
	}); err != nil {
		t.Fatal(err)
	}

	h := NewAuthHandler(service.NewAuthService(repository.NewUserRepository(db)))
	gin.SetMode(gin.TestMode)
	body, _ := json.Marshal(map[string]string{
		"username": "new-user",
		"email":    "new@example.com",
		"password": "password",
	})
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Register(c)

	if rec.Code != 500 {
		t.Fatalf("expected registration failure status, got %d: %s", rec.Code, rec.Body.String())
	}
	var users int64
	if err := db.Model(&domain.User{}).Count(&users).Error; err != nil {
		t.Fatal(err)
	}
	if users != 0 {
		t.Fatalf("registration continued after email check failure")
	}
}
