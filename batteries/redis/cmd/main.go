package main

import (
	"context"

	"github.com/submodule-org/submodule.go"
	"github.com/submodule-org/submodule.go/batteries/redis"
)

type Executable = func() bool

var mylogic = submodule.Make[Executable](func(client *redis.Client) Executable {
	return func() bool {
		return true
	}
}, redis.Mod)

func run(ctx context.Context, args ...string) error {

}

func main() {
	e := mylogic.Resolve()
	e()
}
