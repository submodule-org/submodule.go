package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Addr      string
	Password  string
	Db        int
	isDefault bool
}

type ConfigSetter struct{}

func (c ConfigSetter) WithManually(config Config) *Redis {
	return &Redis{
		config: &config,
	}
}

func (c ConfigSetter) WithDefault() *Redis {
	return &Redis{
		config: &Config{
			Addr:      "localhost:6379",
			Password:  "",
			Db:        0,
			isDefault: true,
		},
	}
}

type Redis struct {
	config *Config
	client *redis.Client
}

func (r *Redis) Connect(ctx context.Context) (*redis.Client, error) {
	r.client = redis.NewClient(&redis.Options{
		Addr:     r.config.Addr,
		Password: r.config.Password,
		DB:       r.config.Db,
	})
	return r.client, nil
}
