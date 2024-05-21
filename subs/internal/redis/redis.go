package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	redis_test_container "github.com/testcontainers/testcontainers-go/modules/redis"
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

func (c ConfigSetter) WithDefault(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Redis, error) {
	if len(opts) == 0 {
		opts = append(opts,
			testcontainers.WithImage("docker.io/redis:7"),
			redis_test_container.WithSnapshotting(10, 1),
			redis_test_container.WithLogLevel(redis_test_container.LogLevelVerbose),
		)
	}

	redisContainer, err := redis_test_container.RunContainer(ctx, opts...)
	if err != nil {
		return nil, err
	}

	ip, _ := redisContainer.Host(ctx)
	port, _ := redisContainer.MappedPort(ctx, "6379/tcp")
	addr := fmt.Sprintf("%s:%s", ip, port.Port())

	return &Redis{
		config: &Config{
			Addr:      addr,
			Password:  "",
			Db:        0,
			isDefault: true,
		},
	}, nil
}

type Redis struct {
	config         *Config
	client         *redis.Client
	redisContainer *redis_test_container.RedisContainer
}

func (r *Redis) Connect(ctx context.Context) (*redis.Client, error) {
	r.client = redis.NewClient(&redis.Options{
		Addr:     r.config.Addr,
		Password: r.config.Password,
		DB:       r.config.Db,
	})
	return r.client, r.client.Ping(ctx).Err()
}

func (r Redis) Close(ctx context.Context) error {
	if r.client == nil && r.redisContainer == nil {
		return nil
	}
	var err error
	if r.client != nil {
		err = r.client.Close()
	}
	if r.redisContainer != nil {
		err = r.redisContainer.Terminate(ctx)
	}
	return err
}
