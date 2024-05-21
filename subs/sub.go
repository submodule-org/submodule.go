package subs

import (
	"github.com/submodule-org/submodule.go"
	"github.com/submodule-org/submodule.go/subs/redis"
)

var RedisSub = submodule.Make[ConfigSetter[redis.Config, redis.Redis]](func() *redis.ConfigSetter {
	return &redis.ConfigSetter{}
})
