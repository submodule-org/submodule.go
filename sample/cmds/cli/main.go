package main

import (
	"os"

	"github.com/submodule-org/submodule.go/v2/meta/mcmd"
	"github.com/submodule-org/submodule.go/v2/sample"
)

func main() {
	mcmd.ResolveCmds(sample.EmptyHandlerRoute)
	mcmd.App.Resolve().Run(os.Args)
}
