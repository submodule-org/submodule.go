// Package submodule offers a simple DI framework without going overboard
//
// # Concept
//
// Each service/components in any systems are likely have dependency on other services/components
// Sometime, those dependencies create an tangible problem. By using just a small chunk of service,
// you instead initiate the whole system.
//
// Submodule was born to solve this problem. Submodule requires you to provide the
// linkage between dependencies. In short, you'll need to define what you need
//
// When a part of system is initializing, submodule will resolve dependencies that needed for the graph.
// By doing so, integration tests become an ease
// You have all benefits of default system wiring, while still refrain from initializing
// the whole system just to test a single service
package submodule

import (
	"fmt"
	"reflect"
)

// `Make` help you create a Submodule from a function
// `Make` input must be a function which
//
// # Any types
//
// With In embedded, all fields of the struct will be resolved
// against dependencies
//
//	func(s *server, l Logger, c Config) (any) {
//	  return nil
//	}
//
// in the example above, s, l and c will be resolved against dependencies
//
// # submodule.In embedded
//
// Will be resolved as a whole against dependencies
//
//	func(p struct {
//	  submodule.In
//	  Server *Server
//	  Logger Logger
//	}, c Config) any {
//		return nil
//	}
//
// in the example above, Server, Logger and Config will be resolved against dependencies
func Make[T any](fn any, dependencies ...Retrievable) Submodule[T] {
	return construct[T](fn, dependencies...)
}

// Resolve fields of struct or struct pointer against given dependencies
//
//	type Server struct {
//	  Logger Logger
//	  Config Config
//	}
//
//	var ServerMod = submodule.Resolve[Server](&Server{}, LoggerMod, ConfigMod)
//
// In the example above, Server.Logger and Server.Config will be resolved against dependencies
func Resolve[T any](t T, dependencies ...Retrievable) Submodule[T] {
	tt := reflect.TypeOf(t)

	if tt.Kind() != reflect.Struct && tt.Kind() != reflect.Pointer {
		panic(fmt.Sprintf("only struct or struct pointer : %v", tt.String()))
	}

	return construct[T](func(self Self) T {
		x, e := resolveEmbedded(self.Scope, tt, reflect.ValueOf(t), self.Dependencies)

		if e != nil {
			panic(e)
		}

		return x.Interface().(T)
	}, dependencies...)
}

// Provide value as it is. Good to setup configurations, keeping types etc
// as is
func Value[T any](t T) Submodule[T] {
	return construct[T](func() T {
		return t
	})
}

// Group groups submodules and re-advertise as a single value
func Group[T any](s ...Retrievable) Submodule[[]T] {
	return construct[[]T](func(self Self) []T {
		var v []T
		for _, submodule := range s {
			t, e := submodule.retrieve(self.Scope)
			if e != nil {
				panic(e)
			}

			v = append(v, t.(T))
		}

		return v
	})
}

// Special type to facitliate dependency injection by struct. Meant to be embed
type In struct{}

// Find all instances of type T within the given scope
func Find[T any](i []T, is Scope) []T {
	t := reflect.TypeOf(i).Elem()
	s := is.(*scope)

	for _, m := range s.values {
		if m.value.Type().AssignableTo(t) {
			i = append(i, m.value.Interface().(T))
		}
	}

	return i
}

// Self is a special type to facitliate dependency injection,
// it will reflect the current dependency list and scope at execution time
type Self struct {
	Scope        Scope
	Dependencies []Retrievable
}
