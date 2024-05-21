package env_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/submodule-org/submodule.go"
	"github.com/submodule-org/submodule.go/batteries/env"
)

func TestEnv(t *testing.T) {
	t.Run("default mode is dev", func(t *testing.T) {
		assert.Equal(t, env.Dev, env.LoadFromEnv())
	})

	t.Run("resolving env works", func(t *testing.T) {
		s := submodule.CreateScope()
		_e, e := env.Mod.SafeResolveWith(s)
		assert.Nil(t, e)
		assert.Equal(t, env.Dev, _e)
	})

	t.Run("can detect env from env var", func(t *testing.T) {
		s := submodule.CreateScope()

		oev := os.Getenv("APP_ENV")
		defer os.Setenv("APP_ENV", oev)

		os.Setenv("APP_ENV", "test")
		_e, e := env.Mod.SafeResolveWith(s)
		assert.Nil(t, e)
		assert.Equal(t, env.Test, _e)
	})

	t.Run("invalid will return to dev", func(t *testing.T) {
		s := submodule.CreateScope()

		oev := os.Getenv("APP_ENV")
		defer os.Setenv("APP_ENV", oev)

		os.Setenv("APP_ENV", "something something")
		_e, e := env.Mod.SafeResolveWith(s)
		assert.Nil(t, e)
		assert.Equal(t, env.Dev, _e)
	})

	t.Run("can change env flag", func(t *testing.T) {
		s := submodule.CreateScope()

		ok := env.EnvKey

		os.Setenv("RANDOM_FLAG", "test")

		env.EnvKey = "RANDOM_FLAG"
		defer func() {
			env.EnvKey = ok
		}()

		_e, e := env.Mod.SafeResolveWith(s)
		assert.Nil(t, e)
		assert.Equal(t, env.Test, _e)
	})

	t.Run("can change load stratgegy", func(t *testing.T) {
		s := submodule.CreateScope()

		ols := env.DefaultStrategy
		env.DefaultStrategy = func() env.Env {
			return env.Prod
		}

		defer func() {
			env.DefaultStrategy = ols
		}()

		_e, e := env.Mod.SafeResolveWith(s)
		assert.Nil(t, e)
		assert.Equal(t, env.Prod, _e)
	})
}
