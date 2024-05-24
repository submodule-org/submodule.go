package main

import "github.com/submodule-org/submodule.go/batteries/sub_cmd"

func main() {
	HelloCmd.Resolve()
	sub_cmd.Start()
}
