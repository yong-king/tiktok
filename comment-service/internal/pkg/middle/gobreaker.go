package middleware

import (
	"github.com/sony/gobreaker"
	"time"
)

// CircuitBreaker 包装了 sony/gobreaker.CircuitBreaker，方便复用。
type CircuitBreaker struct {
	cb *gobreaker.CircuitBreaker
}

// NewCircuitBreaker 创建一个新的熔断器，name 标识熔断器名称。
func NewCircuitBreaker(name string) *CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: 5,                // 半开状态最大允许请求数
		Interval:    60 * time.Second, // 统计窗口周期，计数清零
		Timeout:     30 * time.Second, // 熔断打开后尝试恢复时间
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// 失败率超过60%，且请求数至少10个时打开熔断
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.6
		},
	}
	return &CircuitBreaker{
		cb: gobreaker.NewCircuitBreaker(settings),
	}
}

// Execute 执行函数，如果熔断器打开则直接返回错误。
// 函数内部需要返回 (interface{}, error)，
func (c *CircuitBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {
	return c.cb.Execute(req)
}
