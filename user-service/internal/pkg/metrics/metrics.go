package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"sync"
	"time"
)

var once sync.Once

var (
	// RequestCount 记录请求次数
	RequestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_service_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method", "status"},
	)

	// RequestDuration 记录请求耗时
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "user_service_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method", "status"},
	)

	// 请求错误率
	HTTPErrorCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_service_errors_total",
			Help: "Total number of errors",
		},
		[]string{"path", "method", "status"},
	)

	// DB查询耗时&错误率
	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "user_service_db_query_duration_seconds",
			Help:    "DB query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"query_name"},
	)
	DBQueryErrorCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_service_db_query_errors_total",
			Help: "DB query error count",
		},
		[]string{"query_name", "error"},
	)
)

func Init() {
	// 内存、Goroutine、GC 等 Golang Runtime 指标
	// 可预留后续自定义注册接口（当前使用 promauto 已自动注册）
	once.Do(func() {
		if err := prometheus.Register(collectors.NewGoCollector()); err != nil {
			log.Printf("GoCollector already registered: %v", err)
		}
		if err := prometheus.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
			log.Printf("ProcessCollector already registered: %v", err)
		}
	})
}

// StartMetricsServer 启动独立的 /metrics 服务
func StartMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	log.Println("Prometheus metrics exposed at :18081/metrics")
	go func() {
		if err := http.ListenAndServe(":18081", nil); err != nil {
			log.Fatalf("Metrics server failed: %v", err)
		}
	}()
}

// statusResponseWriter 用来记录返回状态码
type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *statusResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// InstrumentHandler wraps your Kratos HTTP handler to collect Prometheus metrics, InstrumentHandler 是核心中间件
func InstrumentHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusResponseWriter{ResponseWriter: w, statusCode: 200}

		// 执行业务逻辑
		next.ServeHTTP(rw, r)
		duration := time.Since(start).Seconds()

		path := r.URL.Path
		method := r.Method
		status := http.StatusText(rw.statusCode)

		RequestCount.WithLabelValues(path, method, status).Inc()
		RequestDuration.WithLabelValues(path, method, status).Observe(duration)
		if rw.statusCode >= 400 {
			HTTPErrorCount.WithLabelValues(path, method, status).Inc()
		}
	})
}
