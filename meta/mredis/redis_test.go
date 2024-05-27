package mredis_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/submodule-org/submodule.go"
	"github.com/submodule-org/submodule.go/meta/menv"
	"github.com/submodule-org/submodule.go/meta/mredis"
)

func TestRedis(t *testing.T) {

	t.Run("test redis", func(t *testing.T) {
		s := submodule.CreateScope()
		s.InitValue(menv.Mod, menv.Prod)

		client, e := mredis.Mod.SafeResolveWith(s)
		assert.Nil(t, e)

		ctx := context.TODO()

		i, e := client.Info(ctx).Result()
		assert.Nil(t, e)

		fmt.Println(i)
	})

}
