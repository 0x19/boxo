//go:build notracing

package internal

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return noop.NewTracerProvider().Tracer("go-bitswap").Start(ctx, fmt.Sprintf("Bitswap.%s", name), opts...)
}
