package main

import "submodule"

type Config struct {
	Host string
}

var _ = submodule.Provide(func() (Config, error) {
	return Config{
		Host: "localhost",
	}, nil
})
