package core

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/timandy/routine"
)

type In struct{}
type Self struct {
	Dependencies []Gettable
}

var InType = reflect.TypeOf(In{})
var SelfType = reflect.TypeOf(Self{})

type Value[T any] struct {
	mu        sync.Mutex
	value     T
	e         error
	initiated bool
}

type store struct {
	mu     sync.Mutex
	values map[Gettable]*Value[any]
}

func (s *store) Init(g Gettable) *Value[any] {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.values[g]
	if !ok {
		v = &Value[any]{
			initiated: false,
		}
		s.values[g] = v
	}

	return v
}

var localStore = &store{
	values: make(map[Gettable]*Value[any]),
	mu:     sync.Mutex{},
}

var threadLocalStore = routine.NewInheritableThreadLocalWithInitial(func() *store {
	return &store{
		values: make(map[Gettable]*Value[any]),
		mu:     sync.Mutex{},
	}
})

var sandboxFlag = routine.NewInheritableThreadLocalWithInitial[bool](func() bool {
	return false
})

func RunInSandbox(fn func()) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		sandboxFlag.Set(true)
		fn()
	}()
	wg.Wait()
}

func getStore() *store {
	if sandboxFlag.Get() {
		return threadLocalStore.Get()
	}

	return localStore
}

type S[T any] struct {
	Input        any
	ProvideType  reflect.Type
	Dependencies []Gettable
}

type Gettable interface {
	Get() (any, error)
	CanResolve(reflect.Type) bool
}

type Submodule[T any] interface {
	Gettable
	SafeResolve() (T, error)
	Resolve() T

	Init(T)
	Reset()
}

func IsInEmbedded(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Type == InType {
			return true
		}
	}
	return false
}

func IsSelf(t reflect.Type) bool {
	return t.AssignableTo(SelfType)
}

func ResolveEmbedded(t reflect.Type, v reflect.Value, dependencies []Gettable) (reflect.Value, error) {
	var pt reflect.Type
	var pv reflect.Value

	if t.Kind() == reflect.Pointer {
		pv = reflect.Indirect(v)
		pt = t.Elem()
	} else {
		pv = reflect.Indirect(v)
		pt = t
	}

	for i := 0; i < pt.NumField(); i++ {
		f := pt.Field(i)
		if f.Type == InType {
			continue
		}

		if !f.IsExported() {
			return pv, fmt.Errorf("unable to resolve unexported field: %s, field is not exported", f.Name)
		}

		value, err := ResolveType(f.Type, dependencies)
		if err != nil {
			return pv, err
		}

		pv.Field(i).Set(value)
	}

	if t.Kind() == reflect.Pointer {
		return pv.Addr(), nil
	}

	return pv, nil
}

func ResolveType(t reflect.Type, dependencies []Gettable) (v reflect.Value, e error) {
	for _, d := range dependencies {
		if d.CanResolve(t) {
			vv, err := d.Get()
			if err != nil {
				return v, err
			}

			v = reflect.ValueOf(vv)
			return
		}
	}
	return v, fmt.Errorf("unable to resolve dependency for type: %s", t.String())
}

func (s *S[T]) SafeResolve() (t T, e error) {
	store := getStore()

	v := store.Init(s)
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.initiated {
		inputType := reflect.TypeOf(s.Input)

		argsTypes := make([]reflect.Type, inputType.NumIn())
		args := make([]reflect.Value, inputType.NumIn())

		for i := 0; i < inputType.NumIn(); i++ {
			argsTypes[i] = inputType.In(i)

			if IsInEmbedded(argsTypes[i]) {
				var sv reflect.Value
				if argsTypes[i].Kind() == reflect.Pointer {
					sv = reflect.New(argsTypes[i].Elem())
				} else {
					sv = reflect.New(argsTypes[i])
				}

				v, e := ResolveEmbedded(argsTypes[i], sv, s.Dependencies)
				if e != nil {
					return t, fmt.Errorf("unable to resolve embedded type: %s\n %w", argsTypes[i].String(), e)
				}
				args[i] = v
				continue
			}

			if IsSelf(argsTypes[i]) {
				args[i] = reflect.ValueOf(Self{Dependencies: s.Dependencies})
				continue
			}

			v, error := ResolveType(argsTypes[i], s.Dependencies)
			if error != nil {
				return t, error
			}

			args[i] = v
		}

		result := reflect.ValueOf(s.Input).Call(args)
		if len(result) == 1 {
			v.value = result[0].Interface().(T)
		} else {
			if result[1] != reflect.ValueOf(nil) {
				v.e = result[1].Interface().(error)
			} else {
				v.value = result[0].Interface().(T)
			}
		}

		v.initiated = true
	}
	if v.e != nil {
		return t, v.e
	}

	return v.value.(T), v.e
}

func (s *S[T]) Resolve() T {
	r, e := s.SafeResolve()

	if e != nil {
		panic(e)
	}

	return r
}

func (s *S[T]) Get() (any, error) {
	return s.SafeResolve()
}

func (s *S[T]) CanResolve(key reflect.Type) bool {
	return s.ProvideType.AssignableTo(key)
}

func (s *S[T]) Reset() {
	v := getStore().Init(s)
	v.mu.Lock()
	defer v.mu.Unlock()

	v.initiated = false
}

func (s *S[T]) Init(t T) {
	v := getStore().Init(s)
	v.mu.Lock()
	defer v.mu.Unlock()

	v.initiated = true
	v.value = t
	v.e = nil
}

func Construct[T any](
	input any,
	dependencies ...Gettable,
) Submodule[T] {
	inputType := reflect.TypeOf(input)

	if inputType.Kind() != reflect.Func {
		panic(fmt.Sprintf("only func(...any) is accepted, received: %v", inputType.String()))
	}

	provideType := inputType.Out(0)

	if provideType.Kind() == reflect.Interface {
		gt := reflect.TypeOf((*T)(nil)).Elem()
		if !gt.AssignableTo(provideType) {
			panic(fmt.Sprintf("invalid type: %v", provideType))
		}
	} else {
		ot := reflect.New(provideType).Elem().Interface()

		_, ok := ot.(T)
		if !ok {
			panic(fmt.Sprintf("invalid type: %v", ot))
		}
	}

	return &S[T]{
		Input:        input,
		ProvideType:  provideType,
		Dependencies: dependencies,
	}
}
