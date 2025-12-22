package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ctxKey string

// RequestID 中间件：为每个请求添加一个唯一的请求ID
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成一个新的UUID
		reqID := uuid.New().String()
		// 基于已有的请求Context创建一个新的携值Context
		ctx := context.WithValue(c.Request.Context(), ctxKey("request_id"), reqID)
		// 为请求对象绑定新的Context
		c.Request = c.Request.WithContext(ctx)
		// 将request_id设置到响应头
		c.Writer.Header().Set("X-Request-ID", reqID)
		c.Next()
	}
}
