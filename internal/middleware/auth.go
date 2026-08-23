package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"podcast-platform/internal/domain"
	"podcast-platform/internal/service"
	appErr "podcast-platform/pkg/errors"
)

type contextKey string

const (
	CtxUserID   contextKey = "user_id"
	CtxUsername contextKey = "username"
	CtxRole     contextKey = "role"
	CtxClaims   contextKey = "claims"
)

func AuthMiddleware(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResp(appErr.ErrUnauthorized))
			return
		}
		claims, err := authSvc.ValidateToken(tokenStr)
		if err != nil {
			ae, _ := appErr.As(err)
			c.AbortWithStatusJSON(ae.Code, errorResp(ae))
			return
		}
		c.Set(string(CtxUserID), claims.UserID)
		c.Set(string(CtxUsername), claims.Username)
		c.Set(string(CtxRole), claims.Role)
		c.Set(string(CtxClaims), claims)
		c.Next()
	}
}

func OptionalAuthMiddleware(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.Next()
			return
		}
		claims, err := authSvc.ValidateToken(tokenStr)
		if err == nil {
			c.Set(string(CtxUserID), claims.UserID)
			c.Set(string(CtxUsername), claims.Username)
			c.Set(string(CtxRole), claims.Role)
		}
		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(string(CtxRole))
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, errorResp(appErr.ErrForbidden))
			return
		}
		if role != domain.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, errorResp(appErr.ErrForbidden))
			return
		}
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth != "" && strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	if t, err := c.Cookie("token"); err == nil && t != "" {
		return t
	}
	if q := c.Query("token"); q != "" {
		return q
	}
	return ""
}

func GetUserID(c *gin.Context) uint64 {
	if v, ok := c.Get(string(CtxUserID)); ok {
		if id, ok := v.(uint64); ok {
			return id
		}
	}
	return 0
}

func GetUsername(c *gin.Context) string {
	if v, ok := c.Get(string(CtxUsername)); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func GetRole(c *gin.Context) domain.UserRole {
	if v, ok := c.Get(string(CtxRole)); ok {
		if r, ok := v.(domain.UserRole); ok {
			return r
		}
	}
	return ""
}

func IsAdmin(c *gin.Context) bool {
	return GetRole(c) == domain.RoleAdmin
}

func errorResp(err error) gin.H {
	ae, ok := appErr.As(err)
	if ok {
		return gin.H{"code": ae.Code, "message": ae.Message, "error": true}
	}
	return gin.H{"code": 500, "message": err.Error(), "error": true}
}
