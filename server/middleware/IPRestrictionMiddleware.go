/* IP 限制中间件 */
package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var (
	// 黑名单模式
	blackList = map[string]bool{
		"127.0.0.1": true,
	}
	// 白名单模式
	whiteList = map[string]bool{
		"127.0.0.1": true,
	}
)

func IPRestrictionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从环境变量中获取 IP 限制模式
		ipRestrictionMode := os.Getenv("IP_RESTRICTION_MODE")
		if ipRestrictionMode == "" {
			ipRestrictionMode = "blacklist" // 默认使用黑名单模式
		}
		ip := c.ClientIP()
		switch ipRestrictionMode {
		case "blacklist":
			if blackList[ip] {
				c.JSON(http.StatusForbidden, gin.H{
					"status":  "error",
					"message": "IP 权限不足",
				})
				c.Abort()
				return
			}
		case "whitelist":
			if !whiteList[ip] {
				c.JSON(http.StatusForbidden, gin.H{
					"status":  "error",
					"message": "IP 权限不足",
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
