package middleware

import (
	"net/http"

	"github.com/BwCloudWeGo/bw-cli/pkg/redisx"
	"github.com/gin-gonic/gin"
)

type RateLimitConfig struct {
	Enable bool    `mapstructure:"enable" yaml:"enable"` //总开关
	QPS    float64 `mapstructure:"qps" yaml:"qps"`       //每秒允许请求数
	Burst  int     `mapstructure:"burst" yaml:"burst"`   //突发容量
	KeyBy  string  `mapstructure:"key_by" yaml:"key_by"` //
}

func RateLimit(cfg RateLimitConfig, locker *redisx.Locker) gin.HandlerFunc {
	return func(c *gin.Context) {
		if locker == nil {
			c.Next() //如果redis没有初始化就放行
			return
		}
		key := "gw:ratelimit"
		if cfg.KeyBy == "ip" {
			key = "gw:ratelimit:" + c.ClientIP()
		}

		allow, err := locker.Allow(c.Request.Context(), key, int64(cfg.Burst), int64(cfg.QPS))
		if err != nil {
			c.Next()
			return
		}

		if !allow {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": 429, "msg": "Too many requests",
			})
			return
		}
		c.Next()
	}
}
