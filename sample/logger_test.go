package main

import (
	"strings"
	"testing"

	"github.com/submodule-org/submodule.go"
)

func TestLogger(t *testing.T) {
	t.Run("run in info mode should work", func(t *testing.T) {
		submodule.RunInSandbox(func() {
			ConfigMod.Init(Config{
				LogLevel: "info",
			})

			l := LoggerMod.Resolve()
			v := l.Log("test")
			if !strings.HasPrefix(v, "info") {
				t.Fatal()
			}
		})
	})

	t.Run("default logger is debug", func(t *testing.T) {
		submodule.RunInSandbox(func() {
			l := LoggerMod.Resolve()

			v := l.Log("test")
			if !strings.HasPrefix(v, "debug") {
				t.Fatal()
			}
		})
	})

}
