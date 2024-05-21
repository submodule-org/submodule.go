package subs

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

type Suber interface {
	Close(context.Context) error
}

type ConfigSetter[C any, V Suber] interface {
	WithManually(C) *V
	WithDefault(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*V, error)
}
