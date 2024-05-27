package mredis

import (
	"github.com/redis/go-redis/v9"
	"github.com/submodule-org/submodule.go"
	"github.com/submodule-org/submodule.go/meta/mconfig"
)

type Client = redis.Client
type Options = redis.Options

type rawRedisConfig struct {
	Network    string
	Addr       string
	ClientName string
}

var rawRedisConfigMod = mconfig.CreateConfigWithPath("redis", &rawRedisConfig{
	Network:    "tcp",
	Addr:       "127.0.0.1:6379",
	ClientName: "submodule",
})

var configMod = submodule.Make[*Options](func(rc *rawRedisConfig) (*Options, error) {
	return &Options{
		Network:    rc.Network,
		Addr:       rc.Addr,
		ClientName: rc.ClientName,
	}, nil
}, rawRedisConfigMod)

var Mod = submodule.Make[*Client](func(self submodule.Self, config *Options) (*Client, error) {
	client := redis.NewClient(config)

	self.Scope.AppendMiddleware(submodule.WithScopeEnd(func() error {
		return client.Close()
	}))

	return client, nil
}, configMod)
