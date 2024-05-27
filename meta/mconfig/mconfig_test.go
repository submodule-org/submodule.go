package mconfig_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/submodule-org/submodule.go"
	"github.com/submodule-org/submodule.go/meta/mconfig"
)

type TestConfig struct {
	Port       int
	Host       string `mapstruct:"host"`
	Credential struct {
		Username string
		Password string
	} `mapstruct:"credential"`
}

func TestConfigLoader(t *testing.T) {

	t.Run("can load config", func(t *testing.T) {
		mconfig.SetConfigName("test")
		mconfig.SetConfigType("yaml")

		defer mconfig.SetDefaults()

		s := submodule.CreateScope()
		defer s.Dispose()

		loader, e := mconfig.LoaderMod.SafeResolveWith(s)
		assert.Nil(t, e)

		var config = &TestConfig{}
		loader.Load(config)

		assert.Equal(t, 28000, config.Port)
		assert.Equal(t, "admin", config.Credential.Username)

		fmt.Printf("%+v", config)
	})

	t.Run("alter using environment variable", func(t *testing.T) {
		mconfig.SetConfigName("test")
		mconfig.SetConfigType("yaml")

		defer mconfig.SetDefaults()
		ov := os.Getenv("PORT")
		defer os.Setenv("PORT", ov)

		os.Setenv("PORT", "1990")

		s := submodule.CreateScope()
		defer s.Dispose()

		loader, e := mconfig.LoaderMod.SafeResolveWith(s)
		assert.Nil(t, e)

		var config = &TestConfig{}
		loader.Load(config)

		assert.Equal(t, 1990, config.Port)
		assert.Equal(t, "admin", config.Credential.Username)

		fmt.Printf("%+v", config)
	})
}
