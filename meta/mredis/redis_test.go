package mredis_test

import (
	"context"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/submodule-org/submodule.go"
	"github.com/submodule-org/submodule.go/meta/mredis"
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestRedis(t *testing.T) {

	t.Run("test redis", func(t *testing.T) {
		var e error
		ctx := context.TODO()
		container, e := redis.RunContainer(ctx)

		require.Nil(t, e)
		cs, e := container.ConnectionString(ctx)
		log.Printf("connectiong string %s", cs)
		require.Nil(t, e)

		s := submodule.CreateScope()
		mredis.AlterConfig(func(c *mredis.RedisConfig) {
			c.Url = cs
		})

		client, e := mredis.Client.SafeResolveWith(s)
		assert.Nil(t, e)

		i, e := client.Info(ctx).Result()
		assert.Nil(t, e)
		assert.True(t, strings.HasPrefix(i, "# Server"))
	})

}
