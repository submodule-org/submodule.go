package sub_cmd

import (
	"os"

	"github.com/urfave/cli/v2"

	"github.com/submodule-org/submodule.go"
)

type Cmd = *cli.Command

type IntegrateWithUrfave interface {
	AdaptToCLI(app *cli.App)
}

var Mod = submodule.Make[*cli.App](func(self submodule.Self) *cli.App {
	root := &cli.App{}

	cmds := submodule.Find([]IntegrateWithUrfave{}, self.Scope)

	for _, cmd := range cmds {
		cmd.AdaptToCLI(root)
	}

	return root
})

func Start() error {
	root, e := Mod.SafeResolve()
	if e != nil {
		return e
	}

	return root.Run(os.Args)
}

func StartInScope(scope submodule.Scope) error {
	root, e := Mod.SafeResolveWith(scope)
	if e != nil {
		return e
	}

	return root.Run(os.Args)
}
