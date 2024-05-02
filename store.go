package submodule

import (
	"context"
	"reflect"
	"sync"
)

type value struct {
	mu        sync.Mutex
	value     reflect.Value
	e         error
	initiated bool
}

type SubmoduleStore struct {
	mu     sync.Mutex
	values map[Retrievable]*value
}

type Store interface {
	init(g Retrievable) *value
	InitValue(g Retrievable, v any)
	InitError(g Retrievable, e error)
}

func (s *SubmoduleStore) init(g Retrievable) *value {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.values[g]
	if !ok {
		v = &value{
			initiated: false,
		}
		s.values[g] = v
	}

	return v
}

func (s *SubmoduleStore) InitValue(g Retrievable, v any) {
	c := s.init(g)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.value = reflect.ValueOf(v)
	c.initiated = true
}

func (s *SubmoduleStore) InitError(g Retrievable, e error) {
	c := s.init(g)

	c.mu.Lock()
	defer c.mu.Unlock()
	c.e = e
	c.initiated = true
}

func CreateStore() Store {
	return &SubmoduleStore{
		values: make(map[Retrievable]*value),
	}
}

var localStore = CreateStore()

func getStore() Store {
	return localStore
}

type legacyStore struct {
	ctx context.Context
	Store
}

func CreateLegacyStore(ctx context.Context) Store {
	return &legacyStore{
		ctx:   ctx,
		Store: getStore(),
	}
}

func (s *legacyStore) init(g Retrievable) *value {
	if c, ok := s.ctx.Value(g).(Retrievable); ok {
		return s.Store.init(c)
	}

	return s.Store.init(g)
}
