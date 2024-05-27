package main

import (
	"github.com/submodule-org/submodule.go/meta/mcmd"
	"github.com/submodule-org/submodule.go/sample"
)

func main() {
	sample.EmptyHandlerRoute.Resolve()
	mcmd.Start()
}
