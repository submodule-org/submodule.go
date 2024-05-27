package sub_env_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/submodule-org/submodule.go"
	"github.com/submodule-org/submodule.go/meta/menv"
)

func TestEnv(t *testing.T) {
	t.Run("default mode is dev", func(t *testing.T) {
		assert.Equal(t, sub_env.Dev, sub_env.LoadFromEnv())
	})

	t.Run("resolving env works", func(t *testing.T) {
		s := submodule.CreateScope()
		_e, e := sub_env.Mod.SafeResolveWith(s)
		assert.Nil(t, e)
		assert.Equal(t, sub_env.Dev, _e)
	})

	t.Run("can detect env from env var", func(t *testing.T) {
		s := submodule.CreateScope()

		oev := os.Getenv("APP_ENV")
		defer os.Setenv("APP_ENV", oev)

		os.Setenv("APP_ENV", "test")
		_e, e := sub_env.Mod.SafeResolveWith(s)
		assert.Nil(t, e)
		assert.Equal(t, sub_env.Test, _e)
	})

	t.Run("invalid will return to dev", func(t *testing.T) {
		s := submodule.CreateScope()

		oev := os.Getenv("APP_ENV")
		defer os.Setenv("APP_ENV", oev)

		os.Setenv("APP_ENV", "something something")
		_e, e := sub_env.Mod.SafeResolveWith(s)
		assert.Nil(t, e)
		assert.Equal(t, sub_env.Dev, _e)
	})

	t.Run("can change env flag", func(t *testing.T) {
		s := submodule.CreateScope()

		ok := sub_env.EnvKey

		os.Setenv("RANDOM_FLAG", "test")

		sub_env.EnvKey = "RANDOM_FLAG"
		defer func() {
			sub_env.EnvKey = ok
		}()

		_e, e := sub_env.Mod.SafeResolveWith(s)
		assert.Nil(t, e)
		assert.Equal(t, sub_env.Test, _e)
	})

	t.Run("can change load stratgegy", func(t *testing.T) {
		s := submodule.CreateScope()

		ols := sub_env.DefaultStrategy
		sub_env.DefaultStrategy = func() sub_env.Env {
			return sub_env.Prod
		}

		defer func() {
			sub_env.DefaultStrategy = ols
		}()

		_e, e := sub_env.Mod.SafeResolveWith(s)
		assert.Nil(t, e)
		assert.Equal(t, sub_env.Prod, _e)
	})
}
