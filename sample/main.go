package main

import (
	"github.com/submodule-org/submodule.go"
)

func main() {
	// use directly
	gs, e := hiProvider.Resolve()
	if e != nil {
		panic(e)
	}

	gs.Hi()

	// wrap into a function and use
	sayBye := submodule.Factory(func(p struct{ ByeService }) func() string {
		return func() string {
			return p.ByeService.Hi()
		}
	})
	sayBye()

	// wrap into an execution and use
	submodule.Execute(func(p struct{ ByeService }) (any, error) {
		p.ByeService.Bye()
		return nil, nil
	})

}
