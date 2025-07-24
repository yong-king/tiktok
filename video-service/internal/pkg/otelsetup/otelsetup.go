package otelsetup

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc"
	"log"
	"video-service/internal/conf"
)

// InitTracerProvider 初始化并配置全局 OpenTelemetry TraceProvider。
// ctx: 上下文，用于控制超时和退出。
// serviceName: 当前服务名，会在可观测平台（如 Jaeger/Tempo）中展示。
func InitTracerProvider(ctx context.Context, serviceName string, cfg *conf.OpenTelemetry) func(context.Context) error {
	// 定义资源 (Resource)，用于标识服务信息，如 service.name, environment 等。
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName), // 设置服务名称
			attribute.String("env", "dev"),             // 可根据需要设置环境: dev/staging/prod
		),
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	// 创建 OTLP Trace Exporter，通过 gRPC 将链路数据发送到可观测后端 (Jaeger / Tempo)
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(cfg.Endpoint), // Jaeger or Tempo OTLP gRPC endpoint
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	// 创建 BatchSpanProcessor，用于批量发送 Trace 数据（提高性能）
	bsp := sdktrace.NewBatchSpanProcessor(exporter)

	// 创建 TracerProvider，并设置资源信息及 BatchSpanProcessor
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// 设置全局 TracerProvider，之后通过 otel.Tracer("") 获取的 Tracer 即可使用
	otel.SetTracerProvider(tp)

	// 设置全局上下文传播器，TraceContext 用于分布式链路跨服务传递 trace id
	otel.SetTextMapPropagator(propagation.TraceContext{})

	log.Println("🚀 OpenTelemetry tracer initialized.")

	// 返回 shutdown 函数用于优雅关闭（Kratos shutdown 时调用，避免丢失数据）
	return tp.Shutdown
}
