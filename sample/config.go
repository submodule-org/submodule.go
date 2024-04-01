package main

import "submodule"

type Config struct {
	Host string
}

var GetConfig = submodule.Provide(func() (Config, error) {
	return Config{
		Host: "localhost",
	}, nil
})
