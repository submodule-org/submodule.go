package main

import (
	"fmt"

	"github.com/submodule-org/submodule.go"
	"github.com/submodule-org/submodule.go/meta/mconfig"
)

type LoggerConfig struct {
	LogLevel string
}

var LoggerConfigMod = mconfig.CreateConfig(&LoggerConfig{
	LogLevel: "debug",
})

type logger struct {
	Config *LoggerConfig
}

type Logger interface {
	Log(msg string) string
}

func (l *logger) Log(msg string) string {
	return fmt.Sprintf("%s: %s\n", l.Config.LogLevel, msg)
}

var LoggerMod = submodule.Resolve[Logger](&logger{}, LoggerConfigMod)
