package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TraceFunc wraps a function with a span, reducing boilerplate.
func TraceFunc(ctx context.Context, spanName string, fn func(ctx context.Context) error) error {
	tracer := otel.Tracer("app-tracer")
	ctx, span := tracer.Start(ctx, spanName)
	defer span.End()
	return fn(ctx)
}

// StartSpan starts and returns a span with context for more flexible usage.
func StartSpan(ctx context.Context, spanName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	tracer := otel.Tracer("app-tracer")
	ctx, span := tracer.Start(ctx, spanName)
	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}
	return ctx, span
}
