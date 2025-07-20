package server

import (
	middleware "comment-service/internal/pkg/middle"
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer)

// 创建限流中间件，限制全局入口请求
var RateLimitMw = middleware.RateLimitMiddleware(1000, 1888) // 100 QPS，突发200
