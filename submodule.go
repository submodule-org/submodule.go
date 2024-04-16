package submodule

import (
	"context"

	"github.com/submodule-org/submodule.go/internal/core"
)

// ConfigFn is a function that takes a pointer to a Submodule instance and returns a pointer to a Submodule instance.
type Get[T any] core.Get[T]

// Create is a function that takes a factory function and an optional list of configuration functions.
// It returns a pointer to a new Submodule instance.
// The factory function is a function that takes a context and returns a value of any type and an error.
// The configuration functions are used to configure the Submodule instance.
// If the Singleton mode is set in the configuration, the factory function will only be called once.
// The result of the factory function will be cached and returned for all subsequent calls to the Get method of the Submodule.
// If the Singleton mode is not set, the factory function will be called every time the Get method of the Submodule is called.
// @deprecated use Provide, Make or Craft instead
func Create[K any](factory func(context.Context) (K, error), configs ...ConfigFn) Get[K] {
	return Make[K](func() (K, error) {
		ctx := context.Background()
		return factory(ctx)
	})
}

// Derive is a function that takes a factory function, a dependency, and an optional list of configuration functions.
// It returns a Submodule instance.
// The factory function is a function that takes a context and a dependency and returns a value of any type and an error.
// The dependency is an instance that implements the Get interface.
// The configuration functions are used to configure the Submodule instance.
// The factory function will be called every time the Get method of the Submodule is called.
// The result of the factory function will be cached and returned for all subsequent calls to the Get method of the Submodule.
// If the Singleton mode is set in the configuration, the factory function will only be called once.
// If the Singleton mode is not set, the factory function will be called every time the Get method of the Submodule is called.
func Derive[K any, D any, DC Get[D]](
	factory func(context.Context, D) (K, error),
	dep DC,
	configs ...ConfigFn,
) Get[K] {
	wf := func(ctx context.Context) (k K, e error) {
		d, e := dep.Get(ctx)
		if e != nil {
			return
		}

		return factory(ctx, d)
	}

	return Create(wf, configs...)
}

// Execute executes a function with a submodule as a dependency.
func Derive2[V1 any, V2 any, C1 Get[V1], C2 Get[V2], R any](
	factory func(context.Context, V1, V2) (R, error),
	c1 C1, c2 C2,
	configs ...ConfigFn,
) Get[R] {
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

// Derive3 is a function that takes a factory function, three dependencies, and an optional list of configuration functions.
func Derive3[V1 any, V2 any, V3 any, C1 Get[V1], C2 Get[V2], C3 Get[V3], R any](
	factory func(context.Context, V1, V2, V3) (R, error),
	c1 C1, c2 C2, c3 C3,
	configs ...ConfigFn,
) Get[R] {
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
