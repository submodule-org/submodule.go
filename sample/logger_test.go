package main

import (
	"strings"
	"testing"

	"github.com/submodule-org/submodule.go"
)

func TestLogger(t *testing.T) {
	t.Run("run in info mode should work", func(t *testing.T) {
		infoConfig := submodule.Provide(func() Config {
			return Config{
				LogLevel: "info",
			}
		})

		infoLoggerMod := submodule.Prepend(LoggerMod, infoConfig)

		l, e := infoLoggerMod.SafeResolve()

		if e != nil {
			t.Fatalf("Resolve failed %+v", e)
		}

		v := l.Log("test")
		if !strings.HasPrefix(v, "info") {
			t.Fatal()
		}
	})

	t.Run("default logger is debug", func(t *testing.T) {
		l := LoggerMod.Resolve()

		v := l.Log("test")
		if !strings.HasPrefix(v, "debug") {
			t.Fatal()
		}
	})

}
