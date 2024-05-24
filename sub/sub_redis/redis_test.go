package sub_redis_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/submodule-org/submodule.go"
	"github.com/submodule-org/submodule.go/batteries/sub_env"
	"github.com/submodule-org/submodule.go/batteries/sub_redis"
)

func TestRedis(t *testing.T) {

	t.Run("test redis", func(t *testing.T) {
		s := submodule.CreateScope()
		s.InitValue(sub_env.Mod, sub_env.Prod)

		client, e := sub_redis.Mod.SafeResolveWith(s)
		assert.Nil(t, e)

		ctx := context.TODO()

		i, e := client.Info(ctx).Result()
		assert.Nil(t, e)

		fmt.Println(i)
	})

}
