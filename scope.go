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

// A scope is a container for retrievable values.
// To simplify the understanding, a scope is a map, where
// a key is a submodule reference and a value is what will be provided by the factory.
//
// A scope comes with its life-cycle where any submodules can hook to
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

// A scope can enforce a submodule to be a specific value no matter what its factory returns.
// This is useful to simulate test scenarios
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

// A scope can enforce a submodule to be a specific value no matter what its factory returns.
// This is useful to simulate test scenarios
func (s *scope) InitError(g Retrievable, e error) {
	s.initError(g, reflect.ValueOf(e))
}

// Apply middleware to a scope
func (s *scope) Apply(submodule Submodule[Middleware]) {
	m := submodule.ResolveWith(s)

	s.AppendMiddleware(m)
}

// Dispose scope to free up all resolved values and trigger scope end middlewares
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

// Append middleware to the scope
func (s *scope) AppendMiddleware(m Middleware) {
	s.middleware = append(s.middleware, m)
}

// Append global middleware to the global scope
func AppendGlobalMiddleware(ms ...Middleware) {
	for _, m := range ms {
		globalScope.AppendMiddleware(m)
	}
}

// Dispose global scope to free up all resolved values and trigger scope end middlewares
func DisposeGlobalScope() error {
	return globalScope.Dispose()
}

// Apply middleware to the global scope
func Apply(s Middleware) {
	globalScope.AppendMiddleware(s)
}

// A middleware can add behaviors to a scope via decorator pattern.
// There are two types of middlewares
// - a decorator to specific type that will be resolved in the scope
// - a scope end that will be called when the scope is disposed
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

// Create a new scope with modifiers
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

// Return the global scope
func GetStore() Scope {
	return globalScope
}
