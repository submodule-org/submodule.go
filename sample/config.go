package main

import "github.com/submodule-org/submodule.go"

type Config struct {
	Host string
}

var _ = submodule.Provide(func() (Config, error) {
	return Config{
		Host: "localhost",
	}, nil
})
