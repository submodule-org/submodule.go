package sub_env

import (
	"os"

	"github.com/submodule-org/submodule.go"
)

type Env int

func (e Env) IsProd() bool {
	return e == Prod
}

func (e Env) IsNotProd() bool {
	return !e.IsProd()
}

func (e Env) IsTest() bool {
	return e == Test
}

func (e Env) IsNotTest() bool {
	return !e.IsTest()
}

func (e Env) IsDev() bool {
	return e == Dev
}

func (e Env) IsNotDev() bool {
	return !e.IsDev()
}

const (
	Dev Env = iota
	Prod
	Test
)

var EnvKey = "APP_ENV"

func LoadFromEnv() Env {
	env := os.Getenv(EnvKey)

	switch env {
	case "prod":
		return Prod
	case "test":
		return Test
	default:
		return Dev
	}
}

type LoadStrategy func() Env

var DefaultStrategy = LoadFromEnv

var Mod = submodule.Make[Env](func() Env {
	return DefaultStrategy()
})
