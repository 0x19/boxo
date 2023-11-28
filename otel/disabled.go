//go:build !tracing

package otel

import "context"

func Tracer(name string) tracerNoop {
	return tracerNoop{}
}

type tracerNoop struct {
}

func (tracerNoop) Start(ctx context.Context, spanName string, opts ...struct{}) (context.Context, spanNoop) {
	return ctx, spanNoop{}
}

type spanNoop struct{}

func (s spanNoop) End(...any)           {}
func (s spanNoop) RecordError(...any)   {}
func (s spanNoop) AddEvent(...any)      {}
func (s spanNoop) SetStatus(...any)     {}
func (s spanNoop) SetAttributes(...any) {}

func SetTracerProvider(...struct{}) {

}
func SetTextMapPropagator(...any) {

}
