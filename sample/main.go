package main

import "submodule"

func main() {
	gs, e := hiProvider.Resolve()
	if e != nil {
		panic(e)
	}

	gs.Hi()

	sayBye := submodule.Factory(func(p struct{ ByeService }) func() string {
		return func() string {
			return p.ByeService.Hi()
		}
	})

	sayBye()
}
