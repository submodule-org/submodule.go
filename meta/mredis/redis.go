package mredis

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/submodule-org/submodule.go"
	"github.com/submodule-org/submodule.go/meta/menv"
	"github.com/testcontainers/testcontainers-go"
	redisContainer "github.com/testcontainers/testcontainers-go/modules/redis"
)

type Client = redis.Client
type Options = redis.Options

var DefaultOptions = &redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
}

var reusableContainerName = "dev-redis"

var containerMod = submodule.Make[*redisContainer.RedisContainer](func(self submodule.Self, menv menv.Env) *redisContainer.RedisContainer {
	if menv.IsProd() {
		return nil
	}
	ctx := context.Background()
	redisContainer, err := redisContainer.RunContainer(ctx,
		testcontainers.WithImage("docker.io/redis:7"),
		testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
			if menv.IsDev() {
				req.ContainerRequest.Name = reusableContainerName
				req.Reuse = true
			}

			return nil
		}),
		redisContainer.WithSnapshotting(10, 1),
		redisContainer.WithLogLevel(redisContainer.LogLevelVerbose),
	)

	if err != nil {
		log.Fatalf("failed to start container: %s", err)
		panic(err)
	}

	if menv.IsTest() {
		self.Scope.AppendMiddleware(submodule.WithScopeEnd(func() error {
			return redisContainer.Terminate(ctx)
		}))
	}

	return redisContainer
}, menv.Mod)

var configMod = submodule.Make[*Options](func(self submodule.Self, menv menv.Env, container *redisContainer.RedisContainer) (*Options, error) {
	if menv.IsNotProd() {
		ctx := context.Background()
		cs, e := container.ConnectionString(ctx)
		if e != nil {
			return nil, e
		}

		return redis.ParseURL(cs)
	}

	return DefaultOptions, nil
}, menv.Mod, containerMod)

var Mod = submodule.Make[*Client](func(self submodule.Self, config *Options) (*Client, error) {
	client := redis.NewClient(config)

	self.Scope.AppendMiddleware(submodule.WithScopeEnd(func() error {
		return client.Close()
	}))

	return client, nil
}, configMod)
