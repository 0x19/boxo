//go:build tracing

package internal

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer("go-bitswap").Start(ctx, fmt.Sprintf("Bitswap.%s", name), opts...)
}

func StringAttr(k string, v string) trace.SpanStartEventOption {
	return trace.WithAttributes(attribute.String(k, v))
}
