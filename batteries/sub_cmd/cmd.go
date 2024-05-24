package sub_cmd

import (
	"github.com/spf13/cobra"
	"github.com/submodule-org/submodule.go"
)

type Command = cobra.Command

type Cmd interface {
	GetCommand() []*Command
}

var Mod = submodule.Make[*cobra.Command](func(self submodule.Self) *cobra.Command {
	root := &cobra.Command{
		Use: "app",
	}

	cmds := submodule.Find([]Cmd{}, self.Scope)

	for _, cmd := range cmds {
		root.AddCommand(cmd.GetCommand()...)
	}

	return root
})

func Start() error {
	root, e := Mod.SafeResolve()
	if e != nil {
		return e
	}

	return root.Execute()
}

func StartInScope(scope submodule.Scope) error {
	root, e := Mod.SafeResolveWith(scope)
	if e != nil {
		return e
	}

	return root.Execute()
}
