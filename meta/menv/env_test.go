package menv_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/submodule-org/submodule.go"
	"github.com/submodule-org/submodule.go/meta/menv"
)

func TestEnv(t *testing.T) {
	t.Run("default mode is dev", func(t *testing.T) {
		assert.Equal(t, menv.Dev, menv.LoadFromEnv())
	})

	t.Run("resolving env works", func(t *testing.T) {
		s := submodule.CreateScope()
		_e, e := menv.Mod.SafeResolveWith(s)
		assert.Nil(t, e)
		assert.Equal(t, menv.Dev, _e)
	})

	t.Run("can detect env from env var", func(t *testing.T) {
		s := submodule.CreateScope()

		oev := os.Getenv("APP_ENV")
		defer os.Setenv("APP_ENV", oev)

		os.Setenv("APP_ENV", "test")
		_e, e := menv.Mod.SafeResolveWith(s)
		assert.Nil(t, e)
		assert.Equal(t, menv.Test, _e)
	})

	t.Run("invalid will return to dev", func(t *testing.T) {
		s := submodule.CreateScope()

		oev := os.Getenv("APP_ENV")
		defer os.Setenv("APP_ENV", oev)

		os.Setenv("APP_ENV", "something something")
		_e, e := menv.Mod.SafeResolveWith(s)
		assert.Nil(t, e)
		assert.Equal(t, menv.Dev, _e)
	})

	t.Run("can change env flag", func(t *testing.T) {
		s := submodule.CreateScope()

		ok := menv.EnvKey

		os.Setenv("RANDOM_FLAG", "test")

		menv.EnvKey = "RANDOM_FLAG"
		defer func() {
			menv.EnvKey = ok
		}()

		_e, e := menv.Mod.SafeResolveWith(s)
		assert.Nil(t, e)
		assert.Equal(t, menv.Test, _e)
	})

	t.Run("can change load stratgegy", func(t *testing.T) {
		s := submodule.CreateScope()

		ols := menv.DefaultStrategy
		menv.DefaultStrategy = func() menv.Env {
			return menv.Prod
		}

		defer func() {
			menv.DefaultStrategy = ols
		}()

		_e, e := menv.Mod.SafeResolveWith(s)
		assert.Nil(t, e)
		assert.Equal(t, menv.Prod, _e)
	})
}
