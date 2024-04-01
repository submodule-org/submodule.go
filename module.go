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

type Submodule struct {
	store map[reflect.Type]any
}

var defaultStore = &Submodule{
	store: make(map[reflect.Type]any),
}

func Clean() {
	defaultStore.store = make(map[reflect.Type]any)
}

func Show() {
	fmt.Printf("Store %+v\n", defaultStore.store)

}

func Provide[
	V any,
	FN Fn[V],
](fn FN, ms ...[]Submodule) func() (V, error) {
	var once sync.Once
	var instance reflect.Value

	lazyProvider := func() (v V, e error) {
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

	defaultStore.store[reflect.TypeOf(fn).Out(0)] = lazyProvider
	return lazyProvider
}

func Derive[V any, D any, FN DeriveFn[V, D]](fn FN, ms ...[]Submodule) func() (V, error) {
	var once sync.Once
	var instance reflect.Value

	lazyProvider := func() (v V, e error) {
		once.Do(func() {
			providerValue := reflect.ValueOf(fn)
			var args []reflect.Value
			if providerValue.Type().NumIn() > 0 {
				// Assuming a single dependency for simplicity

				depsType := providerValue.Type().In(0)
				dep, error := ResolveByFields(depsType)
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

	defaultStore.store[reflect.TypeOf(fn).Out(0)] = lazyProvider
	return lazyProvider
}

// Generic resolve function
func Resolve(targetType reflect.Type) (v any, e error) {
	p, ok := defaultStore.store[targetType]
	if !ok {
		return v, fmt.Errorf("failed to resolve %s", targetType)
	}

	r := reflect.ValueOf(p).Call(nil)
	fmt.Printf("Resolved %+v \n", r)
	if len(r) != 2 {
		return v, fmt.Errorf("failed to resolve %s", targetType)
	}

	if r[1].Interface() != nil {
		return v, r[1].Interface().(error)
	}

	return r[0].Interface(), nil
}

func ResolveByFields(targetType reflect.Type) (v any, e error) {
	instance := reflect.New(targetType).Elem()
	fields := make([]reflect.StructField, targetType.NumField())

	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		fields[i] = field
		dep, e := Resolve(field.Type)

		if e != nil {
			return v, e
		}

		instance.Field(i).Set(reflect.ValueOf(dep))
	}

	return instance.Interface(), nil
}

func Execute[V any, D any, FN DeriveFn[V, D]](fn FN) (v V, e error) {
	instance, e := ResolveByFields(reflect.TypeOf(fn).In(0))
	if e != nil {
		return
	}

	return fn(instance.(D))
}
