//go:build !tracing

package internal

import (
	"context"
)

func StartSpan(ctx context.Context, name string, opts ...struct{}) (context.Context, FakeSpan) {
	return ctx, spanNoop{}
}

type spanNoop struct{}

func (s spanNoop) End() {}

type FakeSpan interface {
	End()
}

func StringAttr(k string, v string) struct{} {
	return struct{}{}
}

func IntAttr(k string, v int) struct{} {
	return struct{}{}
}
