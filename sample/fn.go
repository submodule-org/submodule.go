package main

import (
	"fmt"

	"github.com/submodule-org/submodule.go"
)

var fn = submodule.Construct(func(p struct {
	submodule.In
	Config
}) func(string) string {
	return func(name string) string {
		return fmt.Sprintf("Hello, %s!", p.Config.Host)
	}
}, ConfigMod)
