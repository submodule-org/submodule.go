package main

import (
	"submodule"
	"testing"
)

func TestGreet(t *testing.T) {
	t.Run("test greet", func(t *testing.T) {
		submodule.Provide(func() (Config, error) {
			return Config{Host: "Test"}, nil
		})

		gs, e := hiProvider.Resolve()
		if e != nil {
			t.Fatalf("Resolve failed %+v", e)
		}

		s := gs.Hi()
		if s != "Test" {
			t.FailNow()
		}
	})
}
