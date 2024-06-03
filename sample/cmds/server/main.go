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

	mhttp.AlterConfig(func(c *mhttp.ServerConfig) {
		c.Addr = ":19000"
	})

	server := mhttp.Server.Resolve()
	e = server.ListenAndServe()
	panic(e)
}
