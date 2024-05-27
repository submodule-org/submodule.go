package main

import (
	"github.com/submodule-org/submodule.go/meta/mhttp"
)

func main() {
	emptyHandlerRoute.Resolve()

	mhttp.Start()
	defer mhttp.Stop()
}
