package mredis

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
	"github.com/submodule-org/submodule.go"
	"github.com/submodule-org/submodule.go/meta/mlogger"
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

var Client = submodule.MakeModifiable[*RedisClient](func(self submodule.Self, config RedisConfig, logger *slog.Logger) (*RedisClient, error) {
	logger.Debug("parsing config", "config object", config)
	opts, e := redis.ParseURL(config.Url)
	if e != nil {
		logger.Error("invalid configuration form", "config", config, slog.Any("error", e))
		return nil, e
	}

	client := redis.NewClient(opts)

	logger.Debug("trying to connect to redis", "url", config.Url)
	e = client.Ping(context.TODO()).Err()

	if e != nil {
		logger.Error("failed to ping redis server, is it connected?", slog.Any("error", e), "url", config.Url)
		return nil, e
	}

	self.Scope.AppendMiddleware(submodule.WithScopeEnd(func() error {
		logger.Info("closing redis client")
		return client.Close()
	}))

	return client, nil
}, defaultRedisConfigMod, mlogger.CreateLogger("redis"))
