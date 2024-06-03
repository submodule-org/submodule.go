package submodule

import (
	"fmt"
	"reflect"
)

var inType = reflect.TypeOf(In{})
var selfType = reflect.TypeOf(Self{})

type submodule[T any] struct {
	input        any
	provideType  reflect.Type
	dependencies []Retrievable
}

type Retrievable interface {
	retrieve(Scope) (any, error)
	canResolve(reflect.Type) bool
}

type Submodule[T any] interface {
	Retrievable
	SafeResolve() (T, error)
	Resolve() T
	ResolveTo(T)

	ResolveWith(Scope) T
	SafeResolveWith(Scope) (T, error)
	ResolveToWith(Scope, T)
}

func (s *submodule[T]) SafeResolve() (t T, e error) {
	return s.SafeResolveWith(nil)
}

func (s *submodule[T]) ResolveWith(as Scope) T {
	t, e := s.SafeResolveWith(as)
	if e != nil {
		panic(e)
	}

	return t
}

func (s *submodule[T]) SafeResolveWith(as Scope) (t T, e error) {
	logger().Debug("resolving",
		"targetType", s.provideType,
		"dependencies", s.dependencies,
	)

	scope := globalScope
	if as != nil {
		scope = as
	}

	var v *value
	if scope.has(s) {
		v = scope.get(s)
		logger().Debug("cache hit", "targetType", s.provideType)
	} else {
		inputType := reflect.TypeOf(s.input)

		argsTypes := make([]reflect.Type, inputType.NumIn())
		args := make([]reflect.Value, inputType.NumIn())

		for i := 0; i < inputType.NumIn(); i++ {
			argsTypes[i] = inputType.In(i)

			if isSelf(argsTypes[i]) {
				args[i] = reflect.ValueOf(Self{
					Scope:        scope,
					Dependencies: s.dependencies,
				})
				continue
			}

			v, err := resolveType(scope, argsTypes[i], s.dependencies)
			if err != nil {
				return t, err
			}

			args[i] = v
		}

		result := reflect.ValueOf(s.input).Call(args)
		if len(result) == 1 {
			v = scope.initValue(s, result[0])
		} else {
			v = scope.initValue(s, result[0])
			if !result[1].IsNil() {
				v.e = result[1]
			}
		}
	}

	if v.e.IsValid() {
		return t, v.e.Interface().(error)
	}

	if v.value.IsZero() {
		return t, e
	}

	return v.value.Interface().(T), nil
}

func (s *submodule[T]) Resolve() T {
	r, e := s.SafeResolve()

	if e != nil {
		panic(e)
	}

	return r
}

func (s *submodule[T]) ResolveTo(t T) {
	s.ResolveToWith(globalScope, t)
}

func (s *submodule[T]) ResolveToWith(as Scope, t T) {
	as.InitValue(s, t)
}

func (s *submodule[T]) retrieve(scope Scope) (any, error) {
	return s.SafeResolveWith(scope)
}

func (s *submodule[T]) canResolve(key reflect.Type) bool {
	return s.provideType.AssignableTo(key)
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

	return &submodule[T]{
		input:        input,
		provideType:  provideType,
		dependencies: dependencies,
	}
}
