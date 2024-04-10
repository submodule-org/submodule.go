package main

import (
	"fmt"

	"github.com/submodule-org/submodule.go"
)

type logger struct {
	Config Config
}

type Logger interface {
	Log(msg string)
}

func (l *logger) Log(msg string) {
	fmt.Printf("%s: %s\n", l.Config.LogLevel, msg)
}

var LoggerMod = submodule.Craft[Logger](&logger{}, ConfigMod)
