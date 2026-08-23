package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"podcast-platform/pkg/logger"
)

var log = logger.New(true)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()
		cost := time.Since(start)
		fields := []zap.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.Int64("duration_ms", cost.Milliseconds()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		}
		if c.Writer.Status() >= 500 {
			log.Error("request", fields...)
		} else if c.Writer.Status() >= 400 {
			log.Warn("request", fields...)
		} else {
			log.Info("request", fields...)
		}
	}
}
