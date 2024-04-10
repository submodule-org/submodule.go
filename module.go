package submodule

import (
	"fmt"
	"reflect"
	"sync"
)

type submodule[T any] struct {
	mu           sync.Mutex
	initiated    bool
	value        T
	e            error
	input        any
	provideType  string
	dependencies []Gettable
}

type Gettable interface {
	Get() (any, error)
	CanResolve(string) bool
}

type Replacable interface {
	Replace(Gettable, Gettable)
}

type Submodule[T any] interface {
	Gettable
	Resolve() (T, error)
}

func (s *submodule[T]) Resolve() (t T, e error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.initiated {
		inputType := reflect.TypeOf(s.input)

		args := make([]reflect.Value, inputType.NumIn())
		for i := 0; i < inputType.NumIn(); i++ {
			paramType := inputType.In(i)

			canResolveField := false
			// loop through dependencies
			for _, d := range s.dependencies {
				if d.CanResolve(paramType.Name()) {
					v, err := d.Get()
					if err != nil {
						e = err
						return
					}

					args[i] = reflect.ValueOf(v)
					canResolveField = true
					break
				}
			}

			if !canResolveField {
				e = fmt.Errorf("unable to resolve dependency for field: %s", paramType.Name())
				return
			}
		}

		result := reflect.ValueOf(s.input).Call(args)
		if len(result) == 1 {
			s.value = result[0].Interface().(T)
		} else {
			if result[1] != reflect.ValueOf(nil) {
				s.e = result[1].Interface().(error)
			} else {
				s.value = result[0].Interface().(T)
			}
		}

		s.initiated = true
	}
	return s.value, s.e
}

func (s *submodule[T]) Get() (any, error) {
	return s.Resolve()
}

func (s *submodule[T]) CanResolve(key string) bool {
	return s.provideType == key
}

func construct[T any](
	input any,
	dependencies ...Gettable,
) Submodule[T] {

	inputType := reflect.TypeOf(input)
	provideType := inputType.Out(0)

	if provideType.Kind() == reflect.Interface {
		fmt.Println("handling interface case")

		gt := reflect.TypeOf((*T)(nil)).Elem()
		if !gt.Implements(provideType) {
			panic(fmt.Sprintf("invalid type: %v", provideType))
		}
	} else {
		ot := reflect.New(provideType).Elem().Interface()

		_, ok := ot.(T)
		if !ok {
			panic(fmt.Sprintf("invalid type: %v", ot))
		}
	}

	return &submodule[T]{
		input:        input,
		provideType:  provideType.Name(),
		dependencies: dependencies,
		initiated:    false,
	}
}

func Provide[T any](fn func() T) Submodule[T] {
	return construct[T](fn)
}

func ProvideWithError[T any](fn func() (T, error)) Submodule[T] {
	return construct[T](fn)
}

func Make[T any](fn any, dependencies ...Gettable) Submodule[T] {
	return construct[T](fn, dependencies...)
}

func Craft[T any](t T) Submodule[T] {
	return construct[T](t)
}

func Override[T any](s Submodule[T], dependencies ...Gettable) {
	sm := s.(*submodule[T])
	var nds []Gettable
	nds = append(nds, dependencies...)
	nds = append(nds, sm.dependencies...)

	sm.dependencies = nds
}
