package main

import (
	"fmt"

	"github.com/submodule-org/submodule.go"
	"github.com/submodule-org/submodule.go/batteries/sub_cmd"
)

type HelloApp struct{}

func (h HelloApp) GetCommand() []*sub_cmd.Command {
	return []*sub_cmd.Command{
		{
			Use: "hello",
			Run: func(cmd *sub_cmd.Command, args []string) {
				fmt.Printf("hello")
			},
		},
	}
}

var HelloCmd = submodule.Resolve(HelloApp{})
