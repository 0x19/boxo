//go:build !tracing

package internal

import (
	"context"
	"fmt"
)

func StartSpan(ctx context.Context, name string, opts ...struct{}) (context.Context, FakeSpan) {
	return ctx, spanNoop{}
}

type spanNoop struct{}

func (s spanNoop) End() {}

type FakeSpan interface {
	End()
}

func StringerAttr(k string, v fmt.Stringer) struct{} {
	return struct{}{}
}
