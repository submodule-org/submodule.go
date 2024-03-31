package submodule

import (
	"context"
	"sync"
)

// ConfigFn is a function that takes a pointer to a Submodule instance and returns a pointer to a Submodule instance.
type Get[V any] interface {
	Get(context.Context) (V, error)
	getOrigin(context.Context) (V, error)
}

// Package submodule provides a set of utilities for managing submodules in a larger project.
// It includes functions for creating, deriving, and executing submodules, as well as a Get interface for retrieving them.
// The package uses context for managing dependencies and error handling.
// It also uses sync for ensuring thread safety.
type Submodule[V any] struct {
	factory func(context.Context) (V, error)
	origin  func(context.Context) (V, error)
}

func (s *Submodule[V]) Get(ctx context.Context) (V, error) {
	select {
	case <-ctx.Done():
		var v V
		return v, ctx.Err()
	default:
		if c, ok := ctx.Value(s).(*Submodule[V]); ok {
			return c.Get(ctx)
		}
		return s.factory(ctx)
	}
}

func (s *Submodule[V]) getOrigin(ctx context.Context) (V, error) {
	select {
	case <-ctx.Done():
		var v V
		return v, ctx.Err()
	default:
		if c, ok := ctx.Value(s).(*Submodule[V]); ok {
			return c.getOrigin(ctx)
		}
		return s.origin(ctx)
	}
}

// Create is a function that takes a factory function and an optional list of configuration functions.
// It returns a pointer to a new Submodule instance.
// The factory function is a function that takes a context and returns a value of any type and an error.
// The configuration functions are used to configure the Submodule instance.
// If the Singleton mode is set in the configuration, the factory function will only be called once.
// The result of the factory function will be cached and returned for all subsequent calls to the Get method of the Submodule.
// If the Singleton mode is not set, the factory function will be called every time the Get method of the Submodule is called.
func Create[K any, C Submodule[K]](factory func(context.Context) (K, error), configs ...ConfigFn) *C {
	var (
		k      K
		e      error
		doOnce sync.Once
	)

	c := buildConfig(configs...)

	wf := factory
	if c.mode == Singleton {
		wf = func(ctx context.Context) (K, error) {
			doOnce.Do(func() {
				k, e = factory(ctx)
			})
			return k, e
		}
	}

	return &C{
		factory: wf,
		origin:  factory,
	}
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
func Derive[K any, D any, C *Submodule[K], DC Get[D]](
	factory func(context.Context, D) (K, error),
	dep DC,
	configs ...ConfigFn,
) C {
	cf := buildConfig(configs...)
	wf := func(ctx context.Context) (k K, e error) {
		if cf.mode == Singleton {
			d, e := dep.Get(ctx)
			if e != nil {
				return k, e
			}

			return factory(ctx, d)
		} else {
			d, e := dep.getOrigin(ctx)
			if e != nil {
				return k, e
			}

			return factory(ctx, d)

		}
	}

	return Create(wf, configs...)
}

// Execute executes a function with a submodule as a dependency.
func Derive2[V1 any, V2 any, C1 Get[V1], C2 Get[V2], R any, RC *Submodule[R]](
	factory func(context.Context, V1, V2) (R, error),
	c1 C1, c2 C2,
	configs ...ConfigFn,
) RC {
	cf := buildConfig(configs...)
	wf := func(ctx context.Context) (r R, e error) {
		if cf.mode == Singleton {
			v1, e := c1.Get(ctx)
			if e != nil {
				return r, e
			}

			v2, e := c2.Get(ctx)
			if e != nil {
				return r, e
			}
			return factory(ctx, v1, v2)
		} else {
			v1, e := c1.getOrigin(ctx)
			if e != nil {
				return r, e
			}

			v2, e := c2.getOrigin(ctx)
			if e != nil {
				return r, e
			}
			return factory(ctx, v1, v2)
		}
	}

	return Create(wf, configs...)
}

// Derive3 is a function that takes a factory function, three dependencies, and an optional list of configuration functions.
func Derive3[V1 any, V2 any, V3 any, C1 Get[V1], C2 Get[V2], C3 Get[V3], R any, RC *Submodule[R]](
	factory func(context.Context, V1, V2, V3) (R, error),
	c1 C1, c2 C2, c3 C3,
	configs ...ConfigFn,
) RC {
	cf := buildConfig(configs...)
	wf := func(ctx context.Context) (r R, e error) {
		if cf.mode == Singleton {
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
		} else {
			v1, e := c1.getOrigin(ctx)
			if e != nil {
				return r, e
			}

			v2, e := c2.getOrigin(ctx)
			if e != nil {
				return r, e
			}

			v3, e := c3.getOrigin(ctx)
			if e != nil {
				return r, e
			}
			return factory(ctx, v1, v2, v3)
		}

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
func Prestage[D any, O any](
	factory func(ctx context.Context, d D) (O, error),
) func(dep Get[D], configs ...ConfigFn) Get[O] {
	return func(dep Get[D], configs ...ConfigFn) Get[O] {
		return Derive(factory, dep, configs...)
	}
}
