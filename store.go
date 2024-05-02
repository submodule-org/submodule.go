package submodule

import (
	"reflect"
	"sync"
)

type value struct {
	mu        sync.Mutex
	value     reflect.Value
	e         error
	initiated bool
}

type Store struct {
	mu     sync.Mutex
	values map[Retrievable]*value
}

func (s *Store) init(g Retrievable) *value {
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

func (s *Store) InitValue(g Retrievable, v any) {
	c := s.init(g)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.value = reflect.ValueOf(v)
	c.initiated = true
}

func (s *Store) InitError(g Retrievable, e error) {
	c := s.init(g)

	c.mu.Lock()
	defer c.mu.Unlock()
	c.e = e
	c.initiated = true
}

func CreateStore() *Store {
	return &Store{
		values: make(map[Retrievable]*value),
	}
}

var localStore = CreateStore()

func getStore() *Store {
	return localStore
}
