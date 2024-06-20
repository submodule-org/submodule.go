package mlogger_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/submodule-org/submodule.go/v2"
	"github.com/submodule-org/submodule.go/v2/meta/mlogger"
)

func TestLogger(t *testing.T) {
	t.Run("run in info mode should work", func(t *testing.T) {
		s := submodule.CreateScope()
		lm := mlogger.CreateLogger("test")

		logger, e := lm.SafeResolveWith(s)
		require.Nil(t, e)
		logger.Info("test")
	})
}
