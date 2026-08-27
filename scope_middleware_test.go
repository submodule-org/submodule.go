package submodule_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/submodule-org/submodule.go/v2"
)

type greeter interface {
	Greet() string
}

type greeterImpl struct {
	msg string
}

func (g greeterImpl) Greet() string {
	return g.msg
}

// Middlewares matching the same type must chain in registration order,
// each one observing the previous one's output.
func Test_Middleware_Chains_In_Order(t *testing.T) {
	var order []string

	first := submodule.WithScopeResolve(func(i int) int {
		order = append(order, "first")
		return i + 1
	})

	second := submodule.WithScopeResolve(func(i int) int {
		order = append(order, "second")
		return i * 10
	})

	mod := submodule.Make[int](func() int { return 1 })
	scope := submodule.CreateScope(submodule.WithMiddlewares(first, second))

	// (1 + 1) * 10 == 20 proves ordering, not just that both ran.
	assert.Equal(t, 20, mod.ResolveWith(scope))
	assert.Equal(t, []string{"first", "second"}, order)
}

// A middleware whose type does not match the resolved value must be skipped.
func Test_Middleware_Skips_Non_Matching_Type(t *testing.T) {
	called := false

	stringMw := submodule.WithScopeResolve(func(s string) string {
		called = true
		return s
	})

	mod := submodule.Make[int](func() int { return 7 })
	scope := submodule.CreateScope(submodule.WithMiddlewares(stringMw))

	assert.Equal(t, 7, mod.ResolveWith(scope))
	assert.False(t, called, "string middleware must not run for an int value")
}

// An interface-typed middleware must keep receiving and returning the
// interface, and a nil return must not panic the scope.
func Test_Middleware_Interface_Typed(t *testing.T) {
	mw := submodule.WithScopeResolve(func(g greeter) greeter {
		return greeterImpl{msg: g.Greet() + "!"}
	})

	mod := submodule.Make[greeter](func() greeter { return greeterImpl{msg: "hi"} })
	scope := submodule.CreateScope(submodule.WithMiddlewares(mw))

	assert.Equal(t, "hi!", mod.ResolveWith(scope).Greet())
}

func Test_Middleware_Nil_Interface_Return(t *testing.T) {
	mw := submodule.WithScopeResolve(func(g greeter) greeter {
		return nil
	})

	mod := submodule.Make[greeter](func() greeter { return greeterImpl{msg: "hi"} })
	scope := submodule.CreateScope(submodule.WithMiddlewares(mw))

	assert.Nil(t, mod.ResolveWith(scope))
}

// A catch-everything func(any) any middleware registered before a typed one
// must not break the typed middleware behind it. The reflect.Value.Call based
// implementation panicked here with "using interface {} as type int".
func Test_Middleware_Widening_Before_Typed(t *testing.T) {
	widen := submodule.WithScopeResolve(func(i any) any { return i })
	typed := submodule.WithScopeResolve(func(i int) int { return i + 100 })

	mod := submodule.Make[int](func() int { return 1 })
	scope := submodule.CreateScope(submodule.WithMiddlewares(widen, typed))

	assert.Equal(t, 101, mod.ResolveWith(scope))
}

// A decorator may declare a wider type than the submodule provides, with
// func(any) any as the common catch-everything case. That must not change how
// the scope stores the value: storing it under the widened interface{} type
// hides it from Find.
func Test_Middleware_Widening_Keeps_Value_Findable(t *testing.T) {
	widen := submodule.WithScopeResolve(func(i any) any { return i })

	mod := submodule.Make[int](func() int { return 3 })

	widened := submodule.CreateScope(submodule.WithMiddlewares(widen))
	assert.Equal(t, 3, mod.ResolveWith(widened))

	plain := submodule.CreateScope()
	assert.Equal(t, 3, mod.ResolveWith(plain))

	assert.Equal(t, []int{3}, submodule.Find([]int{}, plain),
		"without a widening middleware the int is stored as int")
	assert.Equal(t, []int{3}, submodule.Find([]int{}, widened),
		"a widening middleware must not hide the value from Find")
}

// Narrowing back must use the type the submodule declares, not the concrete
// dynamic type, so an interface-providing submodule stays stored as the
// interface.
func Test_Middleware_Widening_Keeps_Interface_Type(t *testing.T) {
	widen := submodule.WithScopeResolve(func(i any) any { return i })

	mod := submodule.Make[greeter](func() greeter { return greeterImpl{msg: "hi"} })
	scope := submodule.CreateScope(submodule.WithMiddlewares(widen))

	assert.Equal(t, "hi", mod.ResolveWith(scope).Greet())
	assert.Len(t, submodule.Find([]greeter{}, scope), 1)
}

// A widening decorator that genuinely replaces the value must still store the
// replacement under the submodule's declared type.
func Test_Middleware_Widening_Replaces_Value(t *testing.T) {
	widen := submodule.WithScopeResolve(func(i any) any { return 99 })

	mod := submodule.Make[int](func() int { return 3 })
	scope := submodule.CreateScope(submodule.WithMiddlewares(widen))

	assert.Equal(t, 99, mod.ResolveWith(scope))
	assert.Equal(t, []int{99}, submodule.Find([]int{}, scope))
}
