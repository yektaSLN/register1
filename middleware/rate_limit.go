// we dont make the rate limit global because in that way the user can block the login by sending 10 requests to products
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	Redis         *redis.Client
	Limit         int
	WindowSeconds int
}

func NewRateLimiter(redisClient *redis.Client, limit int, windowSeconds int) *RateLimiter {
	return &RateLimiter{
		Redis:         redisClient,
		Limit:         limit,
		WindowSeconds: windowSeconds,
	}
}

func (rl *RateLimiter) Allow(ip string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("rate_limit:%s", ip)
	count, err := rl.Redis.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		err := rl.Redis.Expire(ctx, key, time.Duration(rl.WindowSeconds)*time.Second).Err()
		if err != nil {
			return false, err
		}
	}
	if count > int64(rl.Limit) {
		return false, nil
	}

	return true, nil

}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		allowed, err := rl.Allow(ip)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "rate limiter error",
			},
			)
			c.Abort()
			return
		}
		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests",
			},
			)
			c.Abort()
			return
		}
		c.Next()
	}
}
