package sub_cmd

import (
	"os"

	"github.com/urfave/cli/v2"

	"github.com/submodule-org/submodule.go"
)

type Cmd = *cli.Command

type CanHandleCmd interface {
	GetCmd() *cli.Command
}

var Mod = submodule.Make[*cli.App](func(self submodule.Self) *cli.App {
	root := &cli.App{}

	cmds := submodule.Find([]CanHandleCmd{}, self.Scope)

	for _, cmd := range cmds {
		root.Commands = append(root.Commands, cmd.GetCmd())
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
