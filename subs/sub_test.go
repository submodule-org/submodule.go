package subs

import (
	"context"
	"testing"
)

func Test_Redis(t *testing.T) {
	redisSub, err := RedisSub.Resolve().WithDefault(context.Background())
	if err != nil {
		panic(err)
	}
	_, err = redisSub.Connect(context.Background())
	if err != nil {
		panic(err)
	}
	redisSub.Close(context.Background())
}
