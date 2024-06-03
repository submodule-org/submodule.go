package main

import (
	"os"

	"github.com/submodule-org/submodule.go/meta/mcmd"
	"github.com/submodule-org/submodule.go/sample"
)

func main() {
	mcmd.ResolveCmds(sample.EmptyHandlerRoute)
	mcmd.App.Resolve().Run(os.Args)
}
