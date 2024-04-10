package main

import (
	"fmt"

	"github.com/submodule-org/submodule.go"
)

type HelloFn = func(string) string

var fn = submodule.Make[HelloFn](
	func(config Config) HelloFn {
		return func(name string) string {
			return fmt.Sprintf("Hello, %s!", config.Host)
		}
	},
	ConfigMod,
)
