package flow

import (
	"context"

	"github.com/submodule-org/submodule.go"
)

type FlowMiddleware struct {
	submodule.Middleware
}

type FlowContext interface {
	context.Context
	GetFlow() AnyFlow
	GetScope() submodule.Scope
}

type AnyFlow interface {
	resolveActivity(AnyActivity) (any, error)
}

type AnyActivity interface {
	execute(FlowContext) (any, error)
}

type Runnable[T any, R any] interface {
	Run(FlowContext, T) (R, error)
}

type Flow[T any, R any] interface {
	AnyFlow
	Runnable[T, R]
}

type Activity[T any, R any] interface {
	AnyActivity

	Execute(FlowContext, T) (R, error)
}

func Run[T any, R any](ctx context.Context, input submodule.Submodule[Runnable[T, R]], param T) (r R, e error) {
	scope := submodule.CreateScope()
	defer scope.Dispose()

	f, e := input.SafeResolveWith(scope)
	if e != nil {
		return r, e
	}

	fc := flowContext{
		Context: ctx,
		scope:   scope,
	}

	_f := &flow[T, R]{
		ctx:      fc,
		Runnable: f,
	}

	fc.flow = _f

	return f.Run(fc, param)
}

func MakeFlow[T any, R any](
	p any,
	dependencies ...submodule.Retrievable,
) submodule.Submodule[Runnable[T, R]] {
	return submodule.Make[Runnable[T, R]](p, dependencies...)
}

func ResolveFlow[T any, R any](
	p Runnable[T, R],
	dependencies ...submodule.Retrievable,
) submodule.Submodule[Runnable[T, R]] {
	return submodule.Resolve[Runnable[T, R]](p, dependencies...)
}

func Execute[T any, R any](ctx FlowContext, s submodule.Submodule[Executable[T, R]], param T) (r R, e error) {
	a, e := s.SafeResolveWith(ctx.GetScope())
	if e != nil {
		return r, e
	}

	return a.Execute(ctx, param)
}

func MakeActivity[T any, R any](
	p any,
	dependencies ...submodule.Retrievable,
) submodule.Submodule[Executable[T, R]] {
	return submodule.Make[Executable[T, R]](p, dependencies...)
}

func ResolveActivity[T any, R any](
	p Executable[T, R],
	dependencies ...submodule.Retrievable,
) submodule.Submodule[Executable[T, R]] {
	return submodule.Resolve(p, dependencies...)
}
