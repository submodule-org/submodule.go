package mcmd

import (
	"github.com/submodule-org/submodule.go"
	"github.com/urfave/cli/v2"
)

type Cmd = *cli.Command

var App = submodule.Make[*cli.App](func(self submodule.Self) *cli.App {
	root := &cli.App{}

	cmds := submodule.Find([]IntegrateWithUrfave{}, self.Scope)

	for _, cmd := range cmds {
		cmd.AdaptToCLI(root)
	}

	return root
})

func ResolveCmds[T IntegrateWithUrfave](routes ...submodule.Submodule[T]) error {
	return ResolveRoutesIn(submodule.GetStore(), routes...)
}

func ResolveRoutesIn[T IntegrateWithUrfave](scope submodule.Scope, routes ...submodule.Submodule[T]) error {
	for _, r := range routes {
		_, e := r.SafeResolveWith(scope)
		if e != nil {
			return e
		}
	}

	return nil
}
