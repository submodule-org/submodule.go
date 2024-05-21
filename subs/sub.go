package subs

import (
	"github.com/submodule-org/submodule.go"
	"github.com/submodule-org/submodule.go/subs/internal/redis"
)

// check redis sartifies the interface
var _ Suber = (*redis.Redis)(nil)

// check config setter satisfies the interface
var _ ConfigSetter[redis.Config, redis.Redis] = (*redis.ConfigSetter)(nil)

var RedisSub = submodule.Make[ConfigSetter[redis.Config, redis.Redis]](func() *redis.ConfigSetter {
	return &redis.ConfigSetter{}
})
