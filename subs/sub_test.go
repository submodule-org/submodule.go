package subs

import (
	"context"
	"testing"
)

func Test_Redis(t *testing.T) {
	var redisSub = RedisSub.Resolve().WithDefault()
	r, err := redisSub.Connect(context.Background())
	if err != nil {
		panic(err)
	}
	defer r.Close()
}
