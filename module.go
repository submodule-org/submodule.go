package submodule

import (
	"fmt"
	"reflect"

	"github.com/submodule-org/submodule.go/internal/core"
)

type In = core.In

func Provide[T any](fn func() T) core.Submodule[T] {
	return core.Construct[T](fn)
}

func ProvideWithError[T any](fn func() (T, error)) core.Submodule[T] {
	return core.Construct[T](fn)
}

func Make[T any](fn any, dependencies ...core.Gettable) core.Submodule[T] {
	return core.Construct[T](fn, dependencies...)
}

func Craft[T any](t T, dependencies ...core.Gettable) core.Submodule[T] {
	tt := reflect.TypeOf(t)

	if tt.Kind() != reflect.Struct && tt.Kind() != reflect.Pointer {
		panic(fmt.Sprintf("only struct or struct pointer : %v", tt.String()))
	}

	return core.Construct[T](func(self core.Self) T {
		x, e := core.ResolveEmbedded(tt, reflect.ValueOf(t), self.Dependencies)

		if e != nil {
			panic(e)
		}

		return x.Interface().(T)
	}, dependencies...)
}

func Group[T any](s ...core.Gettable) core.Submodule[[]T] {
	return core.Construct[[]T](func() []T {
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

func Prepend[T any](s core.Submodule[T], dependencies ...core.Gettable) core.Submodule[T] {
	osm := s.(*core.S[T])

	var udeps []core.Gettable
	udeps = append(udeps, dependencies...)
	udeps = append(udeps, osm.Dependencies...)

	return core.Construct[T](osm.Input, udeps...)
}
