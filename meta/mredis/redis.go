package mredis

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/submodule-org/submodule.go"
)

type RedisClient = redis.Client
type RedisOptions = redis.Options

type RedisConfig struct {
	Url string
}

var defaultRedisConfig = RedisConfig{
	Url: "redis://localhost:6379",
}

func AlterConfig(c func(*RedisConfig)) {
	mc := &RedisConfig{
		Url: defaultRedisConfig.Url,
	}

	c(mc)
	Client.Append(submodule.Value(*mc))
}

func Reset() {
	Client.Reset()
}

var defaultRedisConfigMod = submodule.Value(defaultRedisConfig)

var Client = submodule.MakeModifiable[*RedisClient](func(self submodule.Self, config RedisConfig) (*RedisClient, error) {
	opts, e := redis.ParseURL(config.Url)
	if e != nil {
		return nil, e
	}

	client := redis.NewClient(opts)
	e = client.Ping(context.TODO()).Err()

	if e != nil {
		return nil, e
	}

	self.Scope.AppendMiddleware(submodule.WithScopeEnd(func() error {
		return client.Close()
	}))

	return client, nil
}, defaultRedisConfigMod)
