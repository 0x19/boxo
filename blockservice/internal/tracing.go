//go:build tracing

package internal

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/attribute"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer("go-blockservice").Start(ctx, fmt.Sprintf("Blockservice.%s", name), opts...)
}

func StringerAttr(k string, v fmt.Stringer) trace.SpanStartEventOption {
	return trace.WithAttributes(attribute.Stringer(k, v))
}
