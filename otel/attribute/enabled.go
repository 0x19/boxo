//go:build tracing

package attribute

import (
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func String(k string, v string) trace.SpanStartEventOption {
	return attribute.String(k, v)
}
func Stringer(k string, v fmt.Stringer) trace.SpanStartEventOption {
	return attribute.Stringer(k, v)
}

func Int(k string, v int) trace.SpanStartEventOption {
	return attribute.Int(k, v)
}
