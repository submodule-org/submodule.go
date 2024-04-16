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

	"github.com/submodule-org/submodule.go/internal/core"
)

// In is the indicator struct to mark a field to be injected
type In = core.In

// RunInSandbox let the submodule to be initiated in a sandbox environment. All initialization will be isolated and will not impact other call. Great for parallel testing
var RunInSandbox = core.RunInSandbox

// Provide help you create a Submodule from a factory function. Provide does not rely on any dependencies
func Provide[T any](fn func() T) core.Submodule[T] {
	return core.Construct[T](fn)
}

// ProvideWithError help you create a Submodule from a factory function that may cause an error
func ProvideWithError[T any](fn func() (T, error)) core.Submodule[T] {
	return core.Construct[T](fn)
}

// `Make` help you create a Submodule from a function
// `Make` input must be a function which
//
// # submodule.In embedded
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
// # Other types
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
func Make[T any](fn any, dependencies ...core.Retrievable) core.Submodule[T] {
	return core.Construct[T](fn, dependencies...)
}

// Resolve fields of struct or struct pointer against given dependencies
//
//	type Server struct {
//	  Logger Logger
//	  Config Config
//	}
//
//	var ServerMod = submodule.Craft[Server](&Server{}, LoggerMod, ConfigMod)
//
// In the example above, Server.Logger and Server.Config will be resolved against dependencies
func Craft[T any](t T, dependencies ...core.Retrievable) core.Submodule[T] {
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

// Group groups submodules and re-advertise as a single value
func Group[T any](s ...core.Retrievable) core.Submodule[[]T] {
	return core.Construct[[]T](func() []T {
		var v []T
		for _, submodule := range s {
			t, e := submodule.Retrieve()
			if e != nil {
				panic(e)
			}

			v = append(v, t.(T))
		}

		return v
	})
}

// Prepend clone the submodule and prepend the dependencies.
// As dependencies are resolved from left to right,
// prepending dependency let you replace certain implementation with others
func Prepend[T any](s core.Submodule[T], dependencies ...core.Retrievable) core.Submodule[T] {
	osm := s.(*core.S[T])

	var updatedDependencies []core.Retrievable
	updatedDependencies = append(updatedDependencies, dependencies...)
	updatedDependencies = append(updatedDependencies, osm.Dependencies...)

	return core.Construct[T](osm.Input, updatedDependencies...)
}
