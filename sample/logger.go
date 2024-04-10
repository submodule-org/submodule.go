package main

import (
	"fmt"

	"github.com/submodule-org/submodule.go"
)

type logger struct {
	config Config
}

type Logger interface {
	Log(msg string)
}

func (l *logger) Log(msg string) {
	fmt.Printf("%s: %s\n", l.config.LogLevel, msg)
}

func createLogger(config Config) Logger {
	return &logger{
		config,
	}
}

var LoggerMod = submodule.Make[Logger](createLogger, ConfigMod)
