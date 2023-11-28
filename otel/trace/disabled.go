//go:build !tracing

package trace

func WithAttributes(attributes ...struct{}) struct{} {
	return struct{}{}
}
