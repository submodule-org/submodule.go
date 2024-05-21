package flow

import (
	"context"

	"github.com/submodule-org/submodule.go"
)

type flowContext struct {
	context.Context
	scope submodule.Scope
	flow  AnyFlow
}

func (f flowContext) GetScope() submodule.Scope {
	return f.scope
}

func (f flowContext) GetFlow() AnyFlow {
	return f.flow
}

type flow[T any, R any] struct {
	Runnable[T, R]
	ctx FlowContext
	fn  func(FlowContext, T) (R, error)
}

func (f *flow[T, R]) resolveActivity(a AnyActivity) (any, error) {
	return a.execute(f.ctx)
}

func (f *flow[T, R]) Run(ctx FlowContext, param T) (R, error) {
	return f.fn(ctx, param)
}
