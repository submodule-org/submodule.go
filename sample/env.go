package main

import "submodule"

type Env struct {
	Name string
}

var _ = submodule.Provide(func() (Env, error) {
	return Env{
		Name: "localhost",
	}, nil
})
