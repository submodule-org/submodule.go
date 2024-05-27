package sample

import (
	"strings"
	"testing"

	"github.com/submodule-org/submodule.go"
)

func TestLogger(t *testing.T) {
	t.Run("run in info mode should work", func(t *testing.T) {
		s := submodule.CreateScope()
		s.InitValue(LoggerConfigMod, &LoggerConfig{
			LogLevel: "info",
		})

		l := LoggerMod.ResolveWith(s)
		v := l.Log("test")
		if !strings.HasPrefix(v, "info") {
			t.Fatal()
		}
	})

	t.Run("default logger is debug", func(t *testing.T) {
		s := submodule.CreateScope()
		l := LoggerMod.ResolveWith(s)

		v := l.Log("test")
		if !strings.HasPrefix(v, "debug") {
			t.Fatal()
		}
	})

}
