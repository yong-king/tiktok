// middleware/ratelimit.go
package middleware

import (
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"golang.org/x/time/rate"
)

// RateLimitMiddleware 返回一个限流中间件，基于令牌桶算法。
// qps 是允许的请求速率，burst 是允许的突发容量。
func RateLimitMiddleware(qps float64, burst int) middleware.Middleware {
	limiter := rate.NewLimiter(rate.Limit(qps), burst)

	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// Allow 允许立即获取一个令牌，如果获取不到直接返回限流错误。
			if !limiter.Allow() {
				return nil, errors.New("too many requests - rate limit exceeded")
			}
			return next(ctx, req)
		}
	}
}
