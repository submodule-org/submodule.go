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

type TestSubPath struct {
	Url string
	Db  int
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

	t.Run("alter nested value using environment variable", func(t *testing.T) {
		mconfig.SetConfigName("test")
		mconfig.SetConfigType("yaml")

		defer mconfig.SetDefaults()
		ov := os.Getenv("CREDENTIAL_USERNAME")
		defer os.Setenv("CREDENTIAL_USERNAME", ov)

		os.Setenv("CREDENTIAL_USERNAME", "user")

		s := submodule.CreateScope()
		defer s.Dispose()

		loader, e := mconfig.LoaderMod.SafeResolveWith(s)
		assert.Nil(t, e)

		var config = &TestConfig{}
		loader.Load(config)

		assert.Equal(t, "user", config.Credential.Username)

		fmt.Printf("%+v", config)
	})

	t.Run("can load by path", func(t *testing.T) {
		defer mconfig.SetDefaults()

		s := submodule.CreateScope()
		defer s.Dispose()

		loader, e := mconfig.LoaderMod.SafeResolveWith(s)
		assert.Nil(t, e)

		var config = &TestSubPath{}
		loader.LoadPath("redis", config)

		assert.Equal(t, 0, config.Db)

		fmt.Printf("%+v", config)
	})

	t.Run("override by path", func(t *testing.T) {
		defer mconfig.SetDefaults()

		os.Setenv("REDIS_DB", "1")

		s := submodule.CreateScope()
		defer s.Dispose()

		loader, e := mconfig.LoaderMod.SafeResolveWith(s)
		assert.Nil(t, e)

		var config = &TestSubPath{}
		loader.LoadPath("redis", config)

		assert.Equal(t, 1, config.Db)

		fmt.Printf("%+v", config)
	})

}
