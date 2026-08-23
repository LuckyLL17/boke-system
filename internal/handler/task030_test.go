package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"podcast-platform/internal/domain"
	"podcast-platform/internal/repository"
	"podcast-platform/internal/service"
)

// PUT /api/v1/auth/profile -> UpdateProfile -> UserRepository.Update -> AuthHandler.UpdateProfile.
func TestProfileUpdateReturnsPersistenceFailure(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:task030?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.User{}); err != nil {
		t.Fatal(err)
	}
	user := &domain.User{Username: "user030", Email: "user030@example.com", PasswordHash: "hash", Role: domain.RoleUser, Status: 1}
	if err := db.Create(user).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Callback().Update().Before("gorm:update").Register("task030_fail_update", func(tx *gorm.DB) {
		tx.Error = errors.New("database write unavailable")
	}); err != nil {
		t.Fatal(err)
	}

	h := NewAuthHandler(service.NewAuthService(repository.NewUserRepository(db)))
	body, _ := json.Marshal(map[string]string{"nickname": "New nickname"})
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Set("user_id", user.ID)
	c.Request = httptest.NewRequest("PUT", "/api/v1/auth/profile", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("profile update panicked on persistence failure")
		}
	}()
	h.UpdateProfile(c)
	if rec.Code != 500 {
		t.Fatalf("expected persistence failure status, got %d: %s", rec.Code, rec.Body.String())
	}
}
