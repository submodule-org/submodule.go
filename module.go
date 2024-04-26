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

type Self = core.Self

// RunInSandbox let the submodule to be initiated in a sandbox environment. All initialization will be isolated and will not impact other call. Great for parallel testing
var RunInSandbox = core.RunInSandbox

// Run let the consumer to execute a function with dependencies
var Run = core.Run

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
//	var ServerMod = submodule.Resolve[Server](&Server{}, LoggerMod, ConfigMod)
//
// In the example above, Server.Logger and Server.Config will be resolved against dependencies
func Resolve[T any](t T, dependencies ...core.Retrievable) core.Submodule[T] {
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
