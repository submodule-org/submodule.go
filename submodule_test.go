package submodule_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/submodule-org/submodule.go/v2"
)

func ms() submodule.Submodule[string] {
	return submodule.Make[string](func() string {
		return "hello"
	})
}

func TestModuleFunction(t *testing.T) {

	t.Run("test module function", func(t *testing.T) {
		s, e := ms().SafeResolve()
		assert.NoError(t, e, "Resolve failed")
		assert.Equal(t, "hello", s, "Unexpected result")

	})

	t.Run("it should fail fast if it knows it can't resolve", func(t *testing.T) {
		x := submodule.Make[int](func() int {
			return 0
		})

		_ = submodule.Make[int](func() int {
			return 1
		})

		_ = submodule.Make[int](func(p struct {
			submodule.In
			A int
		}) int {
			return 1
		}, x)

		assert.Panics(t, func() {
			submodule.Make[int](func(i int) int {
				return 1
			})
		})
	})

	t.Run("test dependency", func(t *testing.T) {
		type A struct {
			Name string
		}
		type B struct {
			Prefix string
		}

		a := submodule.Make[A](func() A {
			return A{
				Name: "hello",
			}
		})

		b := submodule.Make[*B](func(a A) *B {
			return &B{
				Prefix: a.Name,
			}
		}, a)

		xb, e := b.SafeResolve()
		assert.NoError(t, e, "Resolve failed")
		assert.Equal(t, "hello", xb.Prefix, "Unexpected result")

	})

	t.Run("declare wrong type", func(t *testing.T) {
		assert.Panics(t, func() {
			submodule.Make[string](func() int {
				return 0
			})
		})
	})

	t.Run("declare wrong interface", func(t *testing.T) {
		assert.Panics(t, func() {
			submodule.Make[BI](func() AI {
				return As{}
			})
		})
	})

	t.Run("singleton", func(t *testing.T) {
		i := 0

		s := submodule.Make[int](func() int {
			i++
			return i
		})

		x, e := s.SafeResolve()
		assert.NoError(t, e, "Resolve failed")
		assert.Equal(t, 1, x, "Unexpected result")
	})

	t.Run("expose as interface", func(t *testing.T) {
		s := submodule.Make[As](func() AI {
			return As{}
		})

		xs, e := s.SafeResolve()
		assert.NoError(t, e, "Resolve failed")

		xs.Hello()
	})

	t.Run("test craft", func(t *testing.T) {

		a := As{}

		cai := submodule.Resolve[AI](a)
		rcai, e := cai.SafeResolve()
		assert.NoError(t, e, "Resolve failed")
		rcai.Hello()

		cbi := submodule.Resolve[BI](&a)
		rcbi, e := cbi.SafeResolve()
		assert.NoError(t, e, "Resolve failed")
		rcbi.Goodbye()
	})

	t.Run("test In resolve", func(t *testing.T) {

		type A struct {
			Name string
		}

		ma := submodule.Make[A](func() A {
			return A{Name: "hello"}
		})

		mb := submodule.Make[*A](func() *A {
			return &A{Name: "world"}
		})

		a := submodule.Make[string](func(p struct {
			submodule.In
			A  A
			Ap *A
		}) string {
			return p.A.Name
		}, ma, mb)

		s, e := a.SafeResolve()
		assert.NoError(t, e, "Resolve failed")
		assert.Equal(t, "hello", s, "Unexpected result")
	})

	t.Run("group module", func(t *testing.T) {
		type A struct {
			Name string
		}

		a := submodule.Make[A](func() A {
			return A{Name: "hello"}
		})

		b := submodule.Make[A](func() A {
			return A{Name: "world"}
		})

		g := submodule.Group[A](a, b)
		xg, e := g.SafeResolve()

		assert.NoError(t, e, "Resolve failed")
		assert.Equal(t, A{Name: "hello"}, xg[0], "Unexpected result")
		assert.Equal(t, A{Name: "world"}, xg[1], "Unexpected result")
	})

	t.Run("matching interface", func(t *testing.T) {
		a := submodule.Make[As](func() As {
			return As{}
		})

		pa := submodule.Make[*As](func() *As {
			return &As{}
		})

		x := submodule.Make[AI](func(as AI) AI {
			return as
		}, a)

		xa, e := x.SafeResolve()
		assert.Nil(t, e, "Resolve failed")
		assert.NotNil(t, xa, "Resolve failed")

		xa.Hello()

		b := submodule.Make[BI](func(as BI) BI {
			return as
		}, pa)

		xb, e := b.SafeResolve()
		assert.Nil(t, e, "Resolve failed")
		assert.NotNil(t, xb, "Resolve failed")

		xb.Goodbye()

	})

	t.Run("make with error should be fine", func(t *testing.T) {
		me := submodule.Make[int](func() (int, error) {
			return 0, fmt.Errorf("error 2")
		})

		ne := submodule.Make[int](func() (int, error) {
			return 0, nil
		})

		_, e := me.SafeResolve()
		assert.Error(t, e, "Resolve should return error")

		_, e = ne.SafeResolve()
		assert.NoError(t, e, "Resolve should not return error")
	})

	t.Run("error should be treated well", func(t *testing.T) {
		ae := submodule.Make[int](func() (int, error) {
			return 0, fmt.Errorf("error_0")
		})

		_, e := ae.SafeResolve()
		assert.Error(t, e, "Resolve should return error")
		assert.Contains(t, e.Error(), "error_0")

		me := submodule.Make[int](func() (int, error) {
			return 0, fmt.Errorf("error 2")
		})

		_, e = me.SafeResolve()
		assert.Error(t, e, "Resolve should return error")

		ce := submodule.Make[int](func(i int) (int, error) {
			return 0, fmt.Errorf("error 3")
		}, ae)

		_, e = ce.SafeResolve()
		assert.Error(t, e, "Resolve should return error")
		assert.Contains(t, e.Error(), "error_0")
	})

	t.Run("can use isolated store to maintain value", func(t *testing.T) {
		x := submodule.Make[*Counter](func() *Counter {
			return &Counter{
				Count: 0,
			}
		})

		xx := x.Resolve()
		xx.Plus()
		assert.Equal(t, 1, xx.Count, "Count should be 1")

		as := submodule.CreateScope()
		xy := x.ResolveWith(as)
		assert.Equal(t, 0, xy.Count, "Count should be 0")
	})

	t.Run("store can be inherited", func(t *testing.T) {
		x := submodule.Make[*Counter](func() *Counter {
			return &Counter{
				Count: 0,
			}
		})

		xx := x.Resolve()
		xx.Plus()
		assert.Equal(t, 1, xx.Count, "Count should be 1")

		isolatedStore := submodule.CreateScope()
		xy := x.ResolveWith(isolatedStore)
		assert.Equal(t, 0, xy.Count, "Count should be 0")

		inheritedStore := submodule.CreateScope(submodule.Inherit(true))
		xy = x.ResolveWith(inheritedStore)
		assert.Equal(t, 1, xy.Count)

		nestedStore := submodule.CreateScope(submodule.Inherit(true))
		xy = x.ResolveWith(nestedStore)
		assert.Equal(t, 1, xy.Count)

	})

	t.Run("use ResolveTo to init value", func(t *testing.T) {
		x := submodule.Make[*Counter](func() *Counter {
			return &Counter{
				Count: 0,
			}
		})
		var ax *Counter
		var e error

		x.ResolveTo(&Counter{Count: 1})
		ax, e = x.SafeResolve()
		require.Nil(t, e)
		require.Equal(t, 1, ax.Count)

		s := submodule.CreateScope()
		x.ResolveToWith(s, &Counter{Count: 2})
		ax, e = x.SafeResolveWith(s)
		require.Nil(t, e)
		require.Equal(t, 2, ax.Count)
	})

	t.Run("use modifiable submodule", func(t *testing.T) {
		x := submodule.Value(5)

		s := submodule.CreateScope()
		y := submodule.MakeModifiable[int](func(x int) int {
			return x + 1
		}, x)

		z, e := y.SafeResolveWith(s)
		require.Nil(t, e)
		require.Equal(t, 6, z)

		e = s.Dispose()
		require.Nil(t, e)

		y.Append(submodule.Value(7))
		z, e = y.SafeResolveWith(s)
		require.Nil(t, e)
		require.Equal(t, 8, z)
	})

}

type Counter struct {
	Count int
}

func (c *Counter) Plus() {
	c.Count++
}

type As struct{}
type AI interface {
	Hello()
}

type BI interface {
	Goodbye()
}

func (a As) Hello() {
	fmt.Println("hello")
}

func (a *As) Goodbye() {
	fmt.Println("goodbye")
}
