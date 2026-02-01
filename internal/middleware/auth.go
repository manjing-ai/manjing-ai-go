package middleware

import (
	"net/http"
	"strings"
	"time"

	"manjing-ai-go/config"
	"manjing-ai-go/pkg/jwtutil"
	redisclient "manjing-ai-go/pkg/redis"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 鉴权中间件
func AuthMiddleware(cfg config.JWTConfig, rdb *redisclient.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(http.StatusOK, gin.H{"code": 10004, "message": "未授权", "data": gin.H{}})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		if rdb != nil {
			black, err := rdb.IsTokenBlacklisted(c.Request.Context(), tokenStr)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"code": 20001, "message": "系统错误", "data": gin.H{}})
				c.Abort()
				return
			}
			if black {
				c.JSON(http.StatusOK, gin.H{"code": 10004, "message": "未授权", "data": gin.H{}})
				c.Abort()
				return
			}
		}

		claims, err := jwtutil.Parse(tokenStr, cfg)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 10004, "message": "未授权", "data": gin.H{}})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)

		now := time.Now()
		threshold := time.Duration(cfg.RenewThresholdDays) * 24 * time.Hour
		if claims.ExpiresAt != nil && claims.ExpiresAt.Sub(now) <= threshold {
			newToken, err := jwtutil.Generate(claims.UserID, cfg)
			if err == nil {
				c.Header("X-Token", newToken)
			}
		}

		c.Set("token", tokenStr)
		c.Next()
	}
}
