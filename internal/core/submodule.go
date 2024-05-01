package core

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

type Get[V any] interface {
	Get(context.Context) (V, error)
}

type In struct{}
type Self struct {
	Store        *Store
	Dependencies []Retrievable
}

var InType = reflect.TypeOf(In{})
var SelfType = reflect.TypeOf(Self{})

type Value struct {
	mu        sync.Mutex
	value     reflect.Value
	e         error
	initiated bool
}

type Store struct {
	mu     sync.Mutex
	values map[Retrievable]*Value
}

func (s *Store) init(g Retrievable) *Value {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.values[g]
	if !ok {
		v = &Value{
			initiated: false,
		}
		s.values[g] = v
	}

	return v
}

func (s *Store) InitValue(g Retrievable, v any) {
	c := s.init(g)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.value = reflect.ValueOf(v)
	c.initiated = true
}

func (s *Store) InitError(g Retrievable, e error) {
	c := s.init(g)

	c.mu.Lock()
	defer c.mu.Unlock()
	c.e = e
	c.initiated = true
}

func CreateStore() *Store {
	return &Store{
		values: make(map[Retrievable]*Value),
	}
}

var localStore = CreateStore()

func getStore() *Store {
	return localStore
}

type S[T any] struct {
	Input        any
	ProvideType  reflect.Type
	Dependencies []Retrievable
}

type Retrievable interface {
	Retrieve(*Store) (any, error)
	CanResolve(reflect.Type) bool
}

type Submodule[T any] interface {
	Get[T]
	Retrievable
	SafeResolve() (T, error)
	Resolve() T

	ResolveWith(store *Store) T
	SafeResolveWith(store *Store) (T, error)
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

func ResolveEmbedded(as *Store, t reflect.Type, v reflect.Value, dependencies []Retrievable) (reflect.Value, error) {
	var pt reflect.Type
	var pv reflect.Value

	store := getStore()
	if as != nil {
		store = as
	}

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

		value, err := ResolveType(store, f.Type, dependencies)
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

func ResolveType(store *Store, t reflect.Type, dependencies []Retrievable) (v reflect.Value, e error) {
	if IsInEmbedded(t) {
		var sv reflect.Value
		if t.Kind() == reflect.Pointer {
			sv = reflect.New(t.Elem())
		} else {
			sv = reflect.New(t)
		}

		v, e = ResolveEmbedded(store, t, sv, dependencies)
		return
	}

	for _, d := range dependencies {
		if d.CanResolve(t) {
			vv, err := d.Retrieve(store)
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
	return s.SafeResolveWith(nil)
}

func (s *S[T]) ResolveWith(as *Store) T {
	t, e := s.SafeResolveWith(as)
	if e != nil {
		panic(e)
	}

	return t
}

func (s *S[T]) SafeResolveWith(as *Store) (t T, e error) {
	store := getStore()
	if as != nil {
		store = as
	}

	v := store.init(s)
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.initiated {
		inputType := reflect.TypeOf(s.Input)

		argsTypes := make([]reflect.Type, inputType.NumIn())
		args := make([]reflect.Value, inputType.NumIn())

		for i := 0; i < inputType.NumIn(); i++ {
			argsTypes[i] = inputType.In(i)

			if IsSelf(argsTypes[i]) {
				args[i] = reflect.ValueOf(Self{
					Store:        store,
					Dependencies: s.Dependencies,
				})
				continue
			}

			v, error := ResolveType(store, argsTypes[i], s.Dependencies)
			if error != nil {
				return t, error
			}

			args[i] = v
		}

		result := reflect.ValueOf(s.Input).Call(args)
		if len(result) == 1 {
			v.value = result[0]
		} else {
			v.value = result[0]
			if !result[1].IsNil() {
				v.e = result[1].Interface().(error)
			}
		}

		v.initiated = true
	}

	if v.e != nil {
		return t, v.e
	}

	if v.value.IsZero() {
		return t, e
	}

	return v.value.Interface().(T), nil
}

func (s *S[T]) Resolve() T {
	r, e := s.SafeResolve()

	if e != nil {
		panic(e)
	}

	return r
}

func (s *S[T]) Retrieve(store *Store) (any, error) {
	return s.SafeResolveWith(store)
}

func (s *S[T]) CanResolve(key reflect.Type) bool {
	return s.ProvideType.AssignableTo(key)
}

func (s *S[T]) Reset() {
	v := getStore().init(s)
	v.mu.Lock()
	defer v.mu.Unlock()

	v.initiated = false
}

func (s *S[T]) init(t T) {
	v := getStore().init(s)
	v.mu.Lock()
	defer v.mu.Unlock()

	v.initiated = true
	v.value = reflect.ValueOf(t)
	v.e = nil
}

func (s *S[T]) Get(ctx context.Context) (T, error) {
	return s.SafeResolve()
}

func validateInput(input any, isProvider bool) error {
	inputType := reflect.TypeOf(input)

	if inputType.Kind() != reflect.Func {
		return fmt.Errorf("only func(...any) is accepted, received: %v", inputType.String())
	}

	if isProvider {
		if inputType.NumOut() == 0 {
			return fmt.Errorf("provider must return something %v", inputType.String())
		}

		if inputType.NumOut() > 2 {
			return fmt.Errorf("provider must return only one or two values %v", inputType.String())
		}

		if inputType.NumOut() == 2 && !inputType.Out(1).AssignableTo(reflect.TypeOf((*error)(nil)).Elem()) {
			return fmt.Errorf("provider returning a tuple, the 2nd type must be error %v", inputType.String())
		}
	} else {
		if inputType.NumOut() > 1 {
			return fmt.Errorf("run fn can only return none or error %v", inputType.String())
		}

		if inputType.NumOut() == 1 && !inputType.Out(0).AssignableTo(reflect.TypeOf((*error)(nil)).Elem()) {
			return fmt.Errorf("run fn can only return none or error %v", inputType.String())
		}
	}

	return nil
}

func Run(input any, dependencies ...Retrievable) error {
	if err := validateInput(input, false); err != nil {
		return err
	}

	store := getStore()

	runType := reflect.TypeOf(input)
	args := make([]reflect.Value, 0, runType.NumIn())

	for i := 0; i < runType.NumIn(); i++ {
		v, e := ResolveType(store, runType.In(i), dependencies)
		if e != nil {
			fmt.Printf("Resolve failed %+v\n", e)
			return e
		}

		args = append(args, v)
	}

	r := reflect.ValueOf(input).Call(args)

	if len(r) == 1 {
		if !r[0].IsNil() {
			return r[0].Interface().(error)
		}
	}

	return nil
}

func Construct[T any](
	input any,
	dependencies ...Retrievable,
) Submodule[T] {
	inputType := reflect.TypeOf(input)

	if err := validateInput(input, true); err != nil {
		panic(err)
	}

	provideType := inputType.Out(0)

	if provideType.Kind() == reflect.Interface {
		gt := reflect.TypeOf((*T)(nil)).Elem()
		if !gt.AssignableTo(provideType) {
			panic(
				fmt.Sprintf(
					"generic type output mismatch. \n Expect: %s \n Providing: %s",
					gt.String(),
					provideType.String(),
				),
			)
		}
	} else {
		ot := reflect.New(provideType).Elem().Interface()

		_, ok := ot.(T)
		if !ok {
			panic(
				fmt.Sprintf(
					"generic type output mismatch. \n Expect: %s \n Providing: %s",
					ot,
					provideType.String(),
				),
			)
		}
	}

	// check feasibility
	for i := 0; i < inputType.NumIn(); i++ {
		canResolve := false

		pt := inputType.In(i)
		if IsSelf(pt) {
			continue
		}

		if IsInEmbedded(pt) {
			for fi := 0; fi < pt.NumField(); fi++ {
				f := pt.Field(fi)

				if f.Type == InType {
					continue
				}

				for _, d := range dependencies {
					if d.CanResolve(f.Type) {
						canResolve = true
						break
					}
				}

				if !canResolve {
					panic(
						fmt.Sprintf(
							"unable to resolve dependency for type: %s. \n Unable to resolve: %s of %s",
							inputType.String(),
							f.Type.String(),
							pt.String(),
						),
					)
				}
			}
			continue
		}

		for _, d := range dependencies {
			if d.CanResolve(pt) {
				canResolve = true
				break
			}
		}

		if !canResolve {
			panic(
				fmt.Sprintf(
					"unable to resolve dependency for type: %s. \n Unable to resolve: %s",
					inputType.String(),
					pt.String(),
				),
			)
		}
	}

	return &S[T]{
		Input:        input,
		ProvideType:  provideType,
		Dependencies: dependencies,
	}
}
