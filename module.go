package submodule

import (
	"fmt"
	"reflect"
	"sync"
)

type In struct{}

var inType = reflect.TypeOf(In{})

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

func isInEmbed(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Type == inType {
			return true
		}
	}
	return false
}

func resolveEmbedded(st any, dependencies []Gettable) (v any, e error) {
	var t reflect.Type
	var sv reflect.Value

	if reflect.TypeOf(st).Kind() == reflect.Pointer {
		t = reflect.TypeOf(st).Elem()
		sv = reflect.ValueOf(st).Elem()
	} else {
		t = reflect.TypeOf(st)
		sv = reflect.ValueOf(st)
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Type == inType {
			continue
		}

		if !f.IsExported() {
			return v, fmt.Errorf("unable to resolve unexported field: %s", f.Name)
		}

		value, err := resolveType(f.Type, dependencies)
		if err != nil {
			return v, err
		}

		sv.Field(i).Set(value)
	}
	return st, nil
}

func resolveType(t reflect.Type, dependencies []Gettable) (v reflect.Value, e error) {
	for _, d := range dependencies {
		if d.CanResolve(t.Name()) {
			vv, err := d.Get()
			if err != nil {
				return
			}

			v = reflect.ValueOf(vv)
			return
		}
	}
	return v, fmt.Errorf("unable to resolve dependency for type: %s", t.Name())
}

func resolveTypes(types []reflect.Type, dependencies []Gettable) ([]reflect.Value, error) {

	args := make([]reflect.Value, len(types))
	for i := 0; i < len(types); i++ {

		if isInEmbed(types[i]) {
			v, e := resolveEmbedded(types[i], dependencies)
			if e != nil {
				return nil, fmt.Errorf("unable to resolve embedded type: %s, %w", types[i].Name(), e)
			}
			args[i] = reflect.ValueOf(v)
			continue
		}

		v, error := resolveType(types[i], dependencies)
		if error != nil {
			return nil, error
		}

		args[i] = v
	}

	return args, nil
}

func (s *submodule[T]) Resolve() (t T, e error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.initiated {
		inputType := reflect.TypeOf(s.input)

		argsTypes := make([]reflect.Type, inputType.NumIn())
		for i := 0; i < inputType.NumIn(); i++ {
			argsTypes[i] = inputType.In(i)
		}

		args, e := resolveTypes(argsTypes, s.dependencies)
		if e != nil {
			return t, fmt.Errorf("unable to resolve dependencies: %w", e)
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

func Craft[T any](t T, dependencies ...Gettable) Submodule[T] {
	tt := reflect.TypeOf(t)

	if tt.Kind() != reflect.Struct && tt.Kind() != reflect.Pointer && tt.Kind() != reflect.Func {
		panic(fmt.Sprintf("only struct or struct pointer or func: %v", tt.String()))
	}

	return construct[T](func() T {
		_, e := resolveEmbedded(t, dependencies)
		if e != nil {
			e = fmt.Errorf("unable to resolve embedded type: %s, %w", tt.String(), e)
			panic(e)
		}

		return t
	}, dependencies...)
}

func Group[T any](s ...Submodule[T]) Submodule[[]T] {
	return construct[[]T](func() []T {
		var v []T
		for _, submodule := range s {
			t, e := submodule.Get()
			if e != nil {
				panic(e)
			}

			v = append(v, t.(T))
		}

		return v
	})
}

func Override[T any](s Submodule[T], dependencies ...Gettable) {
	sm := s.(*submodule[T])
	var nds []Gettable
	nds = append(nds, dependencies...)
	nds = append(nds, sm.dependencies...)

	sm.dependencies = nds
}
