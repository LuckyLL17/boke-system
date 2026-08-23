package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"podcast-platform/config"
	"podcast-platform/internal/domain"
	"podcast-platform/internal/repository"
	"podcast-platform/internal/service"
)

func TestPasswordChangeRevokesPreviouslyIssuedToken(t *testing.T) {
	db := task010DB(t)
	stamp := time.Now().UnixNano()
	hash, err := bcrypt.GenerateFromPassword([]byte("old-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	user := &domain.User{Username: fmt.Sprintf("user-%d", stamp), Email: fmt.Sprintf("user-%d@example.test", stamp), PasswordHash: string(hash), Status: 1}
	if err := db.Create(user).Error; err != nil {
		t.Fatal(err)
	}
	config.AppConfig = &config.Config{JWT: config.JWTConfig{Secret: "task010-secret", ExpireHours: 1}}
	svc := service.NewAuthService(repository.NewUserRepository(db))
	login, err := svc.Login(&service.LoginRequest{Username: user.Username, Password: "old-password"})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.ChangePassword(user.ID, "old-password", "new-password"); err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", AuthMiddleware(svc), func(c *gin.Context) { c.Status(http.StatusOK) })
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+login.Token)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		panic(fmt.Sprintf("old token remained authorized with status %d", rec.Code))
	}
}

func task010DB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("BOKE_TEST_DSN")
	if dsn == "" {
		dsn = "host=127.0.0.1 port=55432 user=boke password=boke dbname=boke sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.User{}); err != nil {
		t.Fatal(err)
	}
	return db
}
