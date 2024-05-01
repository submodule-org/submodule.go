package submodule_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/submodule-org/submodule.go"
)

func ms() submodule.Submodule[string] {
	return submodule.Make[string](func() string {
		return "hello"
	})
}

func TestModuleFunction(t *testing.T) {

	t.Run("test module function", func(t *testing.T) {
		s, e := ms().SafeResolve()
		if e != nil {
			t.Fatalf("Resolve failed %+v", e)
		}

		if s != "hello" {
			t.FailNow()
		}

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

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("It's expected to be panic")
			}
		}()

		_ = submodule.Make[int](func() int {
			return 0
		})

		_ = submodule.Make[int](func(i int) int {
			return 1
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
		if e != nil {
			t.Fatalf("Resolve failed %+v", e)
		}

		if xb.Prefix != "hello" {
			t.FailNow()
		}

	})

	t.Run("declare wrong type", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("It's expected to be panic")
			}
		}()

		submodule.Make[string](func() int {
			return 0
		})
	})

	t.Run("declare wrong interface", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("It's expected to be panic")
			}
		}()

		submodule.Make[BI](func() AI {
			return As{}
		})
	})

	t.Run("singleton", func(t *testing.T) {
		i := 0

		s := submodule.Make[int](func() int {
			i++
			return i
		})

		_, e := s.SafeResolve()
		if e != nil {
			t.Fatalf("Resolve failed %+v", e)
		}

		ni, _ := s.SafeResolve()

		if ni != 1 {
			fmt.Printf("%+v\n", ni)
			t.FailNow()
		}
	})

	t.Run("expose as interface", func(t *testing.T) {
		s := submodule.Make[As](func() AI {
			return As{}
		})

		xs, e := s.SafeResolve()
		if e != nil {
			t.FailNow()
		}

		xs.Hello()
	})

	t.Run("test craft", func(t *testing.T) {

		a := As{}

		cai := submodule.Resolve[AI](a)
		rcai, e := cai.SafeResolve()

		if e != nil {
			t.Fatalf("Resolve failed %+v", e)
		}
		rcai.Hello()

		cbi := submodule.Resolve[BI](&a)

		rcbi, e := cbi.SafeResolve()

		if e != nil {
			t.Fatalf("Resolve failed %+v", e)
		}
		rcbi.Goodbye()
	})

	t.Run("test In resolve", func(t *testing.T) {

		type A struct {
			Name string
		}

		ma := submodule.Make[A](func() A {
			return A{
				Name: "hello",
			}
		})

		mb := submodule.Make[*A](func() *A {
			return &A{
				Name: "world",
			}
		})

		a := submodule.Make[string](func(p struct {
			submodule.In
			A  A
			Ap *A
		}) string {
			return p.A.Name
		}, ma, mb)

		s, e := a.SafeResolve()
		if e != nil {
			t.Fatalf("Resolve failed %+v", e)
		}

		if s != "hello" {
			t.FailNow()
		}
	})

	t.Run("group module", func(t *testing.T) {
		type A struct {
			Name string
		}

		a := submodule.Make[A](func() A {
			return A{
				Name: "hello",
			}
		})

		b := submodule.Make[A](func() A {
			return A{
				Name: "world",
			}
		})

		g := submodule.Group[A](a, b)
		xg, e := g.SafeResolve()

		if e != nil {
			t.FailNow()
		}

		if xg[0].Name != "hello" || xg[1].Name != "world" {
			fmt.Printf("%+v\n", xg)
			t.FailNow()
		}
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
		if e != nil {
			t.Fatalf("Resolve failed %+v", e)
		}

		xa.Hello()

		b := submodule.Make[BI](func(as BI) BI {
			return as
		}, pa)

		xb, e := b.SafeResolve()
		if e != nil {
			t.Fatalf("Resolve failed %+v", e)
		}

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
		if e == nil {
			t.FailNow()
		}

		_, e = ne.SafeResolve()
		if e != nil {
			t.FailNow()
		}
	})

	t.Run("error should be treated well", func(t *testing.T) {
		ae := submodule.Make[int](func() (int, error) {
			return 0, fmt.Errorf("error_0")
		})

		_, e := ae.SafeResolve()
		if e == nil {
			t.FailNow()
		}

		me := submodule.Make[int](func() (int, error) {
			return 0, fmt.Errorf("error 2")
		})

		_, e = me.SafeResolve()
		if e == nil {
			t.FailNow()
		}

		ce := submodule.Make[int](func(i int) (int, error) {
			return 0, fmt.Errorf("error 3")
		}, ae)

		_, e = ce.SafeResolve()
		if e == nil || !strings.Contains(e.Error(), "error_0") {
			t.FailNow()
		}
	})

	t.Run("can use isolated store to maintain value", func(t *testing.T) {
		x := submodule.Make[*Counter](func() *Counter {
			return &Counter{
				Count: 0,
			}
		})

		xx := x.Resolve()
		xx.Plus()
		if xx.Count != 1 {
			t.Fail()
		}

		as := submodule.CreateStore()
		xy := x.ResolveWith(as)
		if xy.Count != 0 {
			t.Fail()
		}
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
