//go:build tracing

package trace

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func WithAttributes(attributes ...attribute.KeyValue) trace.SpanStartEventOption {
	return trace.WithAttributes(attributes...)
}
