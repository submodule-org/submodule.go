package main

import "github.com/submodule-org/submodule.go"

type Env struct {
	Name string
}

var _ = submodule.Provide(func() (Env, error) {
	return Env{
		Name: "localhost",
	}, nil
})
