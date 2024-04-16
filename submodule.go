package submodule

import (
	"context"

	"github.com/submodule-org/submodule.go/internal/core"
)

// ConfigFn is a function that takes a pointer to a Submodule instance and returns a pointer to a Submodule instance.
type Get[T any] core.Get[T]

// Deprecated: use Provide, Make or Craft instead
func Create[K any](factory func(context.Context) (K, error), configs ...ConfigFn) core.Submodule[K] {
	return Make[K](func() (K, error) {
		ctx := context.Background()
		return factory(ctx)
	})
}

// Deprecated: use Make or Craft instead
func Derive[K any, D any, DC Get[D]](
	factory func(context.Context, D) (K, error),
	dep DC,
	configs ...ConfigFn,
) core.Submodule[K] {
	wf := func(ctx context.Context) (k K, e error) {
		d, e := dep.Get(ctx)
		if e != nil {
			return
		}

		return factory(ctx, d)
	}

	return Create(wf, configs...)
}

// Deprecated: use Provide, Make or Craft instead
func Derive2[V1 any, V2 any, C1 Get[V1], C2 Get[V2], R any](
	factory func(context.Context, V1, V2) (R, error),
	c1 C1, c2 C2,
	configs ...ConfigFn,
) core.Submodule[R] {
	wf := func(ctx context.Context) (r R, e error) {
		v1, e := c1.Get(ctx)
		if e != nil {
			return r, e
		}

		v2, e := c2.Get(ctx)
		if e != nil {
			return r, e
		}

		return factory(ctx, v1, v2)
	}

	return Create(wf, configs...)
}

// Deprecated: use Provide, Make or Craft instead
func Derive3[V1 any, V2 any, V3 any, C1 Get[V1], C2 Get[V2], C3 Get[V3], R any](
	factory func(context.Context, V1, V2, V3) (R, error),
	c1 C1, c2 C2, c3 C3,
	configs ...ConfigFn,
) core.Submodule[R] {
	wf := func(ctx context.Context) (r R, e error) {
		v1, e := c1.Get(ctx)
		if e != nil {
			return r, e
		}

		v2, e := c2.Get(ctx)
		if e != nil {
			return r, e
		}

		v3, e := c3.Get(ctx)
		if e != nil {
			return r, e
		}

		return factory(ctx, v1, v2, v3)
	}

	return Create(wf, configs...)
}

// Deprecated: WIP, will come with a better version later
func Execute[O any, D any, DC Get[D]](
	ctx context.Context,
	executor func(context.Context, D) (O, error),
	dep DC,
) (o O, e error) {
	d, e := dep.Get(ctx)

	if e != nil {
		return o, e
	}

	return executor(ctx, d)
}
