package main

import (
	"github.com/submodule-org/submodule.go/meta/mhttp"
	"github.com/submodule-org/submodule.go/sample"
)

func main() {
	sample.EmptyHandlerRoute.Resolve()

	server := mhttp.Server.Resolve()
	e := server.ListenAndServe()
	panic(e)
}
