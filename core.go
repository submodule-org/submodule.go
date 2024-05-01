package submodule

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
	Store        *store
	Dependencies []Retrievable
}

var inType = reflect.TypeOf(In{})
var selfType = reflect.TypeOf(Self{})

type value struct {
	mu        sync.Mutex
	value     reflect.Value
	e         error
	initiated bool
}

type store struct {
	mu     sync.Mutex
	values map[Retrievable]*value
}

func (s *store) init(g Retrievable) *value {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.values[g]
	if !ok {
		v = &value{
			initiated: false,
		}
		s.values[g] = v
	}

	return v
}

func (s *store) InitValue(g Retrievable, v any) {
	c := s.init(g)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.value = reflect.ValueOf(v)
	c.initiated = true
}

func (s *store) InitError(g Retrievable, e error) {
	c := s.init(g)

	c.mu.Lock()
	defer c.mu.Unlock()
	c.e = e
	c.initiated = true
}

func CreateStore() *store {
	return &store{
		values: make(map[Retrievable]*value),
	}
}

var localStore = CreateStore()

func getStore() *store {
	return localStore
}

type s[T any] struct {
	input        any
	provideType  reflect.Type
	dependencies []Retrievable
}

type Retrievable interface {
	retrieve(*store) (any, error)
	canResolve(reflect.Type) bool
}

type Submodule[T any] interface {
	Get[T]
	Retrievable
	SafeResolve() (T, error)
	Resolve() T

	ResolveWith(store *store) T
	SafeResolveWith(store *store) (T, error)
}

func isInEmbedded(t reflect.Type) bool {
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

func isSelf(t reflect.Type) bool {
	return t.AssignableTo(selfType)
}

func resolveEmbedded(as *store, t reflect.Type, v reflect.Value, dependencies []Retrievable) (reflect.Value, error) {
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
		if f.Type == inType {
			continue
		}

		if !f.IsExported() {
			return pv, fmt.Errorf("unable to resolve unexported field: %s, field is not exported", f.Name)
		}

		value, err := resolveType(store, f.Type, dependencies)
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

func resolveType(store *store, t reflect.Type, dependencies []Retrievable) (v reflect.Value, e error) {
	if isInEmbedded(t) {
		var sv reflect.Value
		if t.Kind() == reflect.Pointer {
			sv = reflect.New(t.Elem())
		} else {
			sv = reflect.New(t)
		}

		v, e = resolveEmbedded(store, t, sv, dependencies)
		return
	}

	for _, d := range dependencies {
		if d.canResolve(t) {
			vv, err := d.retrieve(store)
			if err != nil {
				return v, err
			}

			v = reflect.ValueOf(vv)
			return
		}
	}
	return v, fmt.Errorf("unable to resolve dependency for type: %s", t.String())
}

func (s *s[T]) SafeResolve() (t T, e error) {
	return s.SafeResolveWith(nil)
}

func (s *s[T]) ResolveWith(as *store) T {
	t, e := s.SafeResolveWith(as)
	if e != nil {
		panic(e)
	}

	return t
}

func (s *s[T]) SafeResolveWith(as *store) (t T, e error) {
	store := getStore()
	if as != nil {
		store = as
	}

	v := store.init(s)
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.initiated {
		inputType := reflect.TypeOf(s.input)

		argsTypes := make([]reflect.Type, inputType.NumIn())
		args := make([]reflect.Value, inputType.NumIn())

		for i := 0; i < inputType.NumIn(); i++ {
			argsTypes[i] = inputType.In(i)

			if isSelf(argsTypes[i]) {
				args[i] = reflect.ValueOf(Self{
					Store:        store,
					Dependencies: s.dependencies,
				})
				continue
			}

			v, error := resolveType(store, argsTypes[i], s.dependencies)
			if error != nil {
				return t, error
			}

			args[i] = v
		}

		result := reflect.ValueOf(s.input).Call(args)
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

func (s *s[T]) Resolve() T {
	r, e := s.SafeResolve()

	if e != nil {
		panic(e)
	}

	return r
}

func (s *s[T]) retrieve(store *store) (any, error) {
	return s.SafeResolveWith(store)
}

func (s *s[T]) canResolve(key reflect.Type) bool {
	return s.provideType.AssignableTo(key)
}

func (s *s[T]) Get(ctx context.Context) (T, error) {
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

func run(input any, dependencies ...Retrievable) error {
	if err := validateInput(input, false); err != nil {
		return err
	}

	store := getStore()

	runType := reflect.TypeOf(input)
	args := make([]reflect.Value, 0, runType.NumIn())

	for i := 0; i < runType.NumIn(); i++ {
		v, e := resolveType(store, runType.In(i), dependencies)
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

func construct[T any](
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
		if isSelf(pt) {
			continue
		}

		if isInEmbedded(pt) {
			for fi := 0; fi < pt.NumField(); fi++ {
				f := pt.Field(fi)

				if f.Type == inType {
					continue
				}

				for _, d := range dependencies {
					if d.canResolve(f.Type) {
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
			if d.canResolve(pt) {
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

	return &s[T]{
		input:        input,
		provideType:  provideType,
		dependencies: dependencies,
	}
}
