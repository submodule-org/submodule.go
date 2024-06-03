package submodule

import (
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
	middleware []Middleware
}

type Scope interface {
	get(g Retrievable) *value
	has(g Retrievable) bool

	initValue(g Retrievable, v reflect.Value) *value
	InitValue(g Retrievable, v any)
	initError(g Retrievable, e reflect.Value) *value
	InitError(g Retrievable, e error)

	Dispose() error
	AppendMiddleware(Middleware)
	Apply(Submodule[Middleware])
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

	args := []reflect.Value{v}
	for _, m := range s.middleware {
		if m.hasOnScopeResolve && v.Type().AssignableTo(m.onScopeResolveType) {
			args = m.onScopeResolve.Call(args)
		}
	}

	value := &value{
		value: args[0],
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

func (s *scope) Apply(submodule Submodule[Middleware]) {
	m := submodule.ResolveWith(s)

	s.AppendMiddleware(m)
}

func (s *scope) Dispose() error {
	for _, m := range s.middleware {
		if m.hasOnScopeEnd {
			e := m.onScopeEnd()
			if e != nil {
				return e
			}
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for k := range s.values {
		delete(s.values, k)
	}

	return nil
}

func (s *scope) AppendMiddleware(m Middleware) {
	s.middleware = append(s.middleware, m)
}

func AppendGlobalMiddleware(ms ...Middleware) {
	for _, m := range ms {
		globalScope.AppendMiddleware(m)
	}
}

func DisposeGlobalScope() error {
	return globalScope.Dispose()
}

func Apply(s Submodule[Middleware]) {
	m := s.Resolve()
	globalScope.AppendMiddleware(m)
}

type Middleware struct {
	hasOnScopeResolve bool
	hasOnScopeEnd     bool

	onScopeResolveType reflect.Type
	onScopeResolve     reflect.Value

	onScopeEnd func() error
}

type MiddlewareFn func(Middleware) Middleware

func WithScopeResolve[T any](fn func(T) T) Middleware {
	return Middleware{
		hasOnScopeResolve:  true,
		onScopeResolveType: reflect.TypeOf(fn).In(0),
		onScopeResolve:     reflect.ValueOf(fn),
	}
}

func WithScopeEnd(fn func() error) Middleware {
	return Middleware{
		hasOnScopeEnd: true,
		onScopeEnd:    fn,
	}
}

type ScopeOpts struct {
	inherit     bool
	parent      Scope
	middlewares []Middleware
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

func WithMiddlewares(middlewares ...Middleware) ScopeOptsFn {
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

	s.inherit = opt.inherit
	s.parent = opt.parent

	if len(opt.middlewares) > 0 {
		s.middleware = opt.middlewares
	}

	return s
}

var globalScope = CreateScope(
	Inherit(false),
	WithParent(nil),
)

func GetStore() Scope {
	return globalScope
}
