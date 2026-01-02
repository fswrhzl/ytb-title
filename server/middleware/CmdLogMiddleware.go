/* 基于slog的http请求日志中间件：记录http请求信息，并输出到CMD */
package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func SlogLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		latency := time.Since(start)

		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", latency),
			slog.String("client_ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
		}

		// gin error
		if len(c.Errors) > 0 {
			attrs = append(attrs,
				slog.String("error", c.Errors.String()),
			)
			slog.LogAttrs(c.Request.Context(), slog.LevelError, "http request error", attrs...)
			return
		}

		slog.LogAttrs(c.Request.Context(), slog.LevelInfo, "http request", attrs...)
	}
}
