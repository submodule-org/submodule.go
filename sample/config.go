package main

import (
	"github.com/submodule-org/submodule.go"
)

type Config struct {
	Host     string
	Port     int
	LogLevel string
}

func collectConfig() Config {
	println("resolving config")
	return Config{
		Host:     "localhost",
		Port:     8080,
		LogLevel: "debug",
	}
}

var ConfigMod = submodule.Provide(collectConfig)
