package submodule

import (
	"fmt"
	"reflect"
	"sync"
)

type Kind int

const (
	asIs Kind = iota
	alter
)

type tagger struct {
	tag string
}

type valuer struct {
	mu           sync.Mutex
	initiated    bool
	value        reflect.Value
	e            error
	input        any
	provideType  string
	dependencies []gettable
}

type submodule[T any] struct {
	tagger tagger
	valuer *valuer
	ops    Kind
}

type alterable interface {
	Kind() Kind
	AsGettable() gettable
	AsTaggable() taggable

	gettable
	taggable
}

type gettable interface {
	Get() (any, error)
	CanResolve(string) bool
}

type taggable interface {
	Tag() string
}

type Submodule[T any] interface {
	alterable
	Resolve() (T, error)
}

func (s *submodule[T]) Kind() Kind {
	return s.ops
}

func (s *submodule[T]) AsTaggable() taggable {
	return s
}

func (s *submodule[T]) AsGettable() gettable {
	return s
}

func (s *submodule[T]) Tag() string {
	return s.tagger.tag
}

func (s *submodule[T]) Resolve() (t T, e error) {
	s.valuer.mu.Lock()
	defer s.valuer.mu.Unlock()

	if !s.valuer.initiated {
		inputType := reflect.TypeOf(s.valuer.input)

		argsTypes := make([]reflect.Type, inputType.NumIn())
		for i := 0; i < inputType.NumIn(); i++ {
			argsTypes[i] = inputType.In(i)
		}

		args, e := resolveTypes(argsTypes, s.valuer.dependencies)
		if e != nil {
			return t, fmt.Errorf("unable to resolve dependencies: %w", e)
		}

		result := reflect.ValueOf(s.valuer.input).Call(args)
		if len(result) == 1 {
			s.valuer.value = result[0]
		} else {
			if result[1] != reflect.ValueOf(nil) {
				s.valuer.e = result[1].Interface().(error)
			} else {
				s.valuer.value = result[0]
			}
		}

		s.valuer.initiated = true
	}
	return s.valuer.value.Interface().(T), s.valuer.e
}

func (s *submodule[T]) Get() (any, error) {
	return s.Resolve()
}

func (s *submodule[T]) CanResolve(key string) bool {
	return s.valuer.provideType == key
}

func extract(alterables []alterable) (g []gettable, t []taggable) {
	for _, d := range alterables {
		if d.Kind() == asIs {
			g = append(g, d.AsGettable())
		}

		if d.Kind() == alter {
			t = append(t, d.AsTaggable())
		}
	}

	return
}

func construct[T any](
	input any,
	dependencies ...alterable,
) Submodule[T] {

	inputType := reflect.TypeOf(input)
	if inputType.Kind() != reflect.Func {
		panic(fmt.Sprintf("only func: %v", inputType.String()))
	}

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

	gettables, alters := extract(dependencies)

	var tag string
	for _, a := range alters {
		tag = a.Tag()
	}

	return &submodule[T]{
		ops: asIs,
		tagger: tagger{
			tag: tag,
		},
		valuer: &valuer{
			mu:           sync.Mutex{},
			input:        input,
			provideType:  provideType.Name(),
			dependencies: gettables,
			initiated:    false,
		},
	}
}

func WithTag(s string) alterable {
	return &submodule[any]{
		ops: alter,
		tagger: tagger{
			tag: s,
		},
	}
}
