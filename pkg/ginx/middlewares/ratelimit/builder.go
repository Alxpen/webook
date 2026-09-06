package ratelimit

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

//go:embed slide_window.lua
var slideWindowLua string

var slideWindowScript = redis.NewScript(slideWindowLua)

// Builder 构建基于 Redis 的滑动窗口 IP 限流中间件。
// 同一 Redis 中使用相同前缀的实例共享配额，应保持窗口和阈值一致。
type Builder struct {
	client    *redis.Client
	window    time.Duration
	threshold int
	prefix    string
}

// NewBuilder 限制每个 IP 在任意 window 内最多通过 threshold 次请求。
// client 由调用方管理和关闭；无效配置会在初始化时 panic。
func NewBuilder(client *redis.Client, window time.Duration, threshold int) *Builder {
	if client == nil {
		panic("ratelimit: Redis client must not be nil")
	}
	if window < time.Millisecond || window%time.Millisecond != 0 {
		panic("ratelimit: window must be a positive whole number of milliseconds")
	}
	if threshold <= 0 {
		panic("ratelimit: threshold must be positive")
	}
	return &Builder{client: client, window: window, threshold: threshold, prefix: "ratelimit:ip:"}
}

// Prefix 设置 Redis key 前缀，可用于隔离不同业务或路由组的配额。
// 应在 Build 前调用，且不要并发修改 Builder。
func (b *Builder) Prefix(prefix string) *Builder {
	if prefix == "" {
		panic("ratelimit: prefix must not be empty")
	}
	b.prefix = prefix
	return b
}

// Build 返回 Gin 中间件：超限返回 429，Redis 异常返回 503。
// IP 来自 ctx.ClientIP()，调用方应通过 Gin.SetTrustedProxies 配置信任的代理。
func (b *Builder) Build() gin.HandlerFunc {
	client, window, threshold, prefix := b.client, b.window.Milliseconds(), b.threshold, b.prefix
	return func(ctx *gin.Context) {
		// 使用随机请求标识，避免同一毫秒内的并发请求覆盖 ZSET 成员。
		var requestID [16]byte
		if _, err := rand.Read(requestID[:]); err != nil {
			_ = ctx.Error(err)
			ctx.AbortWithStatus(http.StatusServiceUnavailable)
			return
		}
		retryAfter, err := func() (int64, error) {
			// 同时限制连接池等待和 Redis 执行耗时，并继承请求取消信号。
			redisCtx, cancel := context.WithTimeout(ctx.Request.Context(), time.Second)
			defer cancel()
			return slideWindowScript.Run(redisCtx, client,
				[]string{prefix + ctx.ClientIP()}, window, threshold,
				hex.EncodeToString(requestID[:])).Int64()
		}()
		if err != nil {
			_ = ctx.Error(err)
			ctx.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"message": "限流服务暂不可用"})
			return
		}
		if retryAfter > 0 {
			ctx.Header("Retry-After", strconv.FormatInt((retryAfter+999)/1000, 10))
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"message": "请求过于频繁，请稍后重试"})
			return
		}
		ctx.Next()
	}
}
