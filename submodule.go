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
func Derive[K any, D any](
	factory func(context.Context, D) (K, error),
	dep core.Submodule[D],
	configs ...ConfigFn,
) core.Submodule[K] {
	return Make[K](func(d D) (K, error) {
		ctx := context.Background()
		return factory(ctx, d)
	}, dep)
}

// Deprecated: use Provide, Make or Craft instead
func Derive2[V1 any, V2 any, R any](
	factory func(context.Context, V1, V2) (R, error),
	c1 core.Submodule[V1], c2 core.Submodule[V2],
	configs ...ConfigFn,
) core.Submodule[R] {
	return Make[R](func(v1 V1, v2 V2) (R, error) {
		ctx := context.Background()
		return factory(ctx, v1, v2)
	}, c1, c2)
}

// Deprecated: use Provide, Make or Craft instead
func Derive3[V1 any, V2 any, V3 any, R any](
	factory func(context.Context, V1, V2, V3) (R, error),
	c1 core.Submodule[V1], c2 core.Submodule[V2], c3 core.Submodule[V3],
	configs ...ConfigFn,
) core.Submodule[R] {
	return Make[R](func(v1 V1, v2 V2, v3 V3) (R, error) {
		ctx := context.Background()
		return factory(ctx, v1, v2, v3)
	}, c1, c2, c3)
}

// Deprecated: WIP, will come with a better version later
func Execute[O any, D any](
	ctx context.Context,
	executor func(context.Context, D) (o O, e error),
	dep core.Submodule[D],
) (o O, e error) {
	e = core.Run(func(d D) error {
		ao, e := executor(ctx, d)
		o = ao

		return e
	}, dep)

	if e != nil {
		return o, e
	}

	return o, nil
}
