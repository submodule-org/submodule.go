package main

import (
	"github.com/submodule-org/submodule.go/meta/mhttp"
	"github.com/submodule-org/submodule.go/sample"
)

func main() {
	e := mhttp.ResolveRoutes(sample.EmptyHandlerRoute)
	if e != nil {
		panic(e)
	}

	server := mhttp.Server.Resolve()
	e = server.ListenAndServe()
	panic(e)
}
