package submodule

import (
	"context"
	"reflect"
	"sync"
)

type value struct {
	value reflect.Value
	e     reflect.Value
}

type scope struct {
	mu     sync.Mutex
	values map[Retrievable]*value

	parent     Scope
	inherit    bool
	middleware []ScopeMiddleware
}

type Scope interface {
	get(g Retrievable) *value
	has(g Retrievable) bool
	initValue(g Retrievable, v reflect.Value) *value
	InitValue(g Retrievable, v any)
	initError(g Retrievable, e reflect.Value) *value
	InitError(g Retrievable, e error)
	Dispose()
}

func (s *scope) has(g Retrievable) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.inherit && globalScope.has(g) {
		return true
	}

	if s.parent != nil && s.parent.has(g) {
		return true
	}

	_, ok := s.values[g]
	return ok
}

func (s *scope) get(g Retrievable) *value {
	s.mu.Lock()
	defer s.mu.Unlock()

	var v *value
	var ok bool

	if s.parent != nil && s.parent.has(g) {
		v = s.parent.get(g)
		s.values[g] = v

		return v
	}

	if s.inherit && globalScope.has(g) {
		return globalScope.get(g)
	}

	v, ok = s.values[g]
	if !ok {
		g.retrieve(s)
		v = s.values[g]
	}

	return v
}

func (s *scope) initValue(g Retrievable, v reflect.Value) *value {
	if s.has(g) {
		return s.get(g)
	}

	value := &value{
		value: v,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[g] = value

	return value
}

func (s *scope) InitValue(g Retrievable, v any) {
	s.initValue(g, reflect.ValueOf(v))

}

func (s *scope) initError(g Retrievable, e reflect.Value) *value {
	if s.has(g) {
		return s.get(g)
	}

	value := &value{
		e: e,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[g] = value

	return value
}

func (s *scope) InitError(g Retrievable, e error) {
	s.initError(g, reflect.ValueOf(e))
}

func (s *scope) Dispose() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k := range s.values {
		delete(s.values, k)
	}
}

type ScopeMiddleware struct {
	onScopeEnd func(Scope)
}

type ScopeOpts struct {
	inherit     bool
	parent      Scope
	middlewares []ScopeMiddleware
}

type ScopeOptsFn func(opts ScopeOpts) ScopeOpts

func MakeScopeOpts(fn ScopeOptsFn) ScopeOpts {
	return fn(ScopeOpts{})
}

func Inherit(inherit bool) ScopeOptsFn {
	return func(opts ScopeOpts) ScopeOpts {
		opts.inherit = inherit
		return opts
	}
}

func WithParent(parent Scope) ScopeOptsFn {
	return func(opts ScopeOpts) ScopeOpts {
		opts.parent = parent
		return opts
	}
}

func WithMiddlewares(middlewares ...ScopeMiddleware) ScopeOptsFn {
	return func(opts ScopeOpts) ScopeOpts {
		opts.middlewares = middlewares
		return opts
	}
}

func CreateScope(fns ...ScopeOptsFn) Scope {
	s := &scope{
		values: make(map[Retrievable]*value),
	}

	opt := ScopeOpts{}
	for _, fn := range fns {
		opt = fn(opt)
	}

	if opt.inherit {
		s.inherit = true
	}

	if opt.parent != nil {
		s.parent = opt.parent
	}

	if len(opt.middlewares) > 0 {
		s.middleware = opt.middlewares
	}

	return s
}

var globalScope = CreateScope(
	Inherit(false),
	WithParent(nil),
)

func getStore() Scope {
	return globalScope
}

type legacyScope struct {
	ctx context.Context
	Scope
}

var storeMapLock sync.Mutex
var legacyScopeMap = make(map[context.Context]Scope)

func CreateLegacyStore(ctx context.Context) Scope {
	if s, ok := legacyScopeMap[ctx]; ok {
		return s
	}

	store := &legacyScope{
		ctx:   ctx,
		Scope: CreateScope(),
	}

	storeMapLock.Lock()
	defer storeMapLock.Unlock()
	legacyScopeMap[ctx] = store

	return store
}

func DisposeLegacyScope(ctx context.Context) {
	storeMapLock.Lock()
	defer storeMapLock.Unlock()
	delete(legacyScopeMap, ctx)
}

func (s *legacyScope) get(g Retrievable) *value {
	if s.ctx.Value(g) != nil {
		if c, ok := s.ctx.Value(g).(Retrievable); ok {
			v, e := c.retrieve(s.Scope)
			if e != nil {
				return &value{
					e: reflect.ValueOf(e),
				}
			}

			return &value{
				value: reflect.ValueOf(v),
			}
		}

		return &value{
			value: reflect.ValueOf(s.ctx.Value(g)),
		}
	}

	return s.Scope.get(g)
}

func (s *legacyScope) has(g Retrievable) bool {
	if s.Scope.has(g) {
		return true
	}

	return s.ctx.Value(g) != nil
}
