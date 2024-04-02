package submodule

import (
	"fmt"
	"reflect"
	"sync"
)

type Fn[V any] interface {
	func() (V, error)
}

type DeriveFn[V any, D any] interface {
	func(D) (V, error)
}

type Provider[V any, FN Fn[V]] struct {
	factory func() (any, error)
	store   map[reflect.Type]func() (any, error)
	Submodule
}

func (p *Provider[V, FN]) Resolve() (v V, e error) {
	r, e := p.Get()
	if e != nil {
		return v, e
	}

	return r.(V), nil
}

func (p *Provider[V, FN]) Store(key reflect.Type, value func() (any, error)) {
	p.store[key] = value
}

func (p *Provider[V, FN]) Get() (any, error) {
	return p.factory()
}

func (p *Provider[V, FN]) CanResolve(key reflect.Type) bool {
	_, ok := p.store[key]
	return ok
}

type Submodule interface {
	Store(key reflect.Type, value func() (any, error))
	Get() (any, error)
	CanResolve(key reflect.Type) bool
}

var DefaultStore = make(map[reflect.Type]func() (any, error))

func defaultStore(key reflect.Type, value func() (any, error)) {
	DefaultStore[key] = value
}

func defaultResolve(key reflect.Type) (any, error) {
	if defaultCanResolve(key) {
		return DefaultStore[key]()
	}

	return nil, fmt.Errorf("cannot resolve %s", key)
}

func defaultCanResolve(key reflect.Type) bool {
	_, ok := DefaultStore[key]
	return ok
}

func Provide[
	V any,
	FN Fn[V],
](fn FN) *Provider[V, FN] {
	var once sync.Once
	var instance reflect.Value

	lazyProvider := func() (v any, e error) {
		once.Do(func() {
			providerValue := reflect.ValueOf(fn)
			var args []reflect.Value
			r := providerValue.Call(args)

			if len(r) != 2 {
				e = fmt.Errorf("failed to resolve %s", providerValue.Type())
				return
			}

			if r[1].Interface() != nil {
				e = r[1].Interface().(error)
				return
			}

			instance = r[0]

		})

		return instance.Interface().(V), nil
	}

	provider := &Provider[V, FN]{
		factory: lazyProvider,
		store:   make(map[reflect.Type]func() (any, error)),
	}

	provider.Store(reflect.TypeOf(fn).Out(0), lazyProvider)
	defaultStore(reflect.TypeOf(fn).Out(0), lazyProvider)

	return provider
}

func Derive[
	V any,
	D any,
	FN DeriveFn[V, D],
	RFN Fn[V],
](fn FN, ms ...Submodule) *Provider[V, RFN] {
	var once sync.Once
	var instance reflect.Value

	lazyProvider := func() (v any, e error) {
		once.Do(func() {
			providerValue := reflect.ValueOf(fn)
			var args []reflect.Value
			if providerValue.Type().NumIn() > 0 {
				// Assuming a single dependency for simplicity

				depsType := providerValue.Type().In(0)
				dep, error := resolveByFields(depsType, ms...)
				fmt.Printf("> Resolved r: %+v e: %+v\n", dep, error)
				if error != nil {
					e = error
					return
				}

				args = append(args, reflect.ValueOf(dep))
			}
			instance = providerValue.Call(args)[0]
		})

		if e != nil {
			fmt.Printf("> Resolved r: %+v e: %+v\n", v, e)
			return v, e
		}

		return instance.Interface().(V), e
	}

	provider := &Provider[V, RFN]{
		factory: lazyProvider,
		store:   make(map[reflect.Type]func() (any, error)),
	}

	provider.Store(reflect.TypeOf(fn).Out(0), lazyProvider)
	defaultStore(reflect.TypeOf(fn).Out(0), lazyProvider)

	return provider
}

func ResolveByType[T any](st T, ms ...Submodule) (t T, e error) {
	v, e := resolve(reflect.TypeOf(st), ms...)
	if e != nil {
		return t, e
	}

	t, ok := v.(T)
	if !ok {
		return t, fmt.Errorf("cannot convert resolved value to type %s", reflect.TypeOf(st))
	}
	return t, nil
}

/**
 * resolve function
 * One of those submodules will give the resolved value of the target type
 * return error if it cannot resolve
 */
// Generic resolve function
func resolve(targetType reflect.Type, ms ...Submodule) (v any, e error) {
	for _, m := range ms {
		if m.CanResolve(targetType) {
			return m.Get()
		}
	}

	if defaultCanResolve(targetType) {
		return defaultResolve(targetType)
	}

	return v, fmt.Errorf("cannot resolve %s", targetType)
}

func resolveByFields(targetType reflect.Type, ms ...Submodule) (v any, e error) {
	instance := reflect.New(targetType).Elem()
	fields := make([]reflect.StructField, targetType.NumField())

	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		fields[i] = field
		dep, e := resolve(field.Type, ms...)

		if e != nil {
			return v, e
		}

		instance.Field(i).Set(reflect.ValueOf(dep))
	}

	return instance.Interface(), nil
}

func Execute[V any, D any, FN DeriveFn[V, D]](fn FN) (v V, e error) {
	instance, e := resolveByFields(reflect.TypeOf(fn).In(0))
	if e != nil {
		return
	}

	return fn(instance.(D))
}

func Factory[S any, V any](fn func(S) V) (v V) {
	// check if v is function, if not panic
	tv := reflect.TypeOf(fn)
	if tv.Out(0).Kind() != reflect.Func {
		panic("Factory function must return a function")
	}

	vType := tv.Out(0)

	// create a reflection-based function to replace v
	rfn := reflect.MakeFunc(tv.Out(0), func(args []reflect.Value) []reflect.Value {
		instance, e := resolveByFields(reflect.TypeOf(fn).In(0))
		if e != nil {
			if vType.NumOut() == 2 {
				return []reflect.Value{reflect.ValueOf(nil), reflect.ValueOf(e)}
			}

			panic("failed to handle error")
		}

		v := fn(instance.(S))
		rv := reflect.ValueOf(v)

		return rv.Call(args)
	})

	return rfn.Interface().(V)
}
