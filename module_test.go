package submodule

import (
	"fmt"
	"strings"
	"testing"

	"github.com/submodule-org/submodule.go/internal/core"
)

func ms() core.Submodule[string] {
	return Make[string](func() string {
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

	t.Run("test dependency", func(t *testing.T) {
		type A struct {
			Name string
		}
		type B struct {
			Prefix string
		}

		a := Make[A](func() A {
			return A{
				Name: "hello",
			}
		})

		b := Make[*B](func(a A) *B {
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

		Make[string](func() int {
			return 0
		})
	})

	t.Run("declare wrong interface", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("It's expected to be panic")
			}
		}()

		Make[BI](func() AI {
			return As{}
		})
	})

	t.Run("overriding", func(t *testing.T) {
		type A struct {
			Name string
		}

		type B struct {
			Prefix string
		}

		a := Make[A](func() A {
			return A{
				Name: "hello",
			}
		})

		aa := Make[A](func() A {
			return A{
				Name: "world",
			}
		})

		b := Make[B](func(a A) B {
			return B{
				Prefix: a.Name + "hello",
			}
		}, a)

		mb := Prepend(b, aa)
		xb := mb.Resolve()

		if xb.Prefix != "worldhello" {
			fmt.Printf("%+v\n", xb)
			t.FailNow()
		}
	})

	t.Run("singleton", func(t *testing.T) {
		i := 0

		s := Make[int](func() int {
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
		s := Make[As](func() AI {
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

		cai := Craft[AI](a)
		rcai, e := cai.SafeResolve()

		if e != nil {
			t.Fatalf("Resolve failed %+v", e)
		}
		rcai.Hello()

		cbi := Craft[BI](&a)

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

		ma := Provide(func() A {
			return A{
				Name: "hello",
			}
		})

		mb := Provide(func() *A {
			return &A{
				Name: "world",
			}
		})

		a := Make[string](func(p struct {
			core.In
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

		a := Make[A](func() A {
			return A{
				Name: "hello",
			}
		})

		b := Make[A](func() A {
			return A{
				Name: "world",
			}
		})

		g := Group[A](a, b)
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
		a := Make[As](func() As {
			return As{}
		})

		pa := Make[*As](func() *As {
			return &As{}
		})

		x := Make[AI](func(as AI) AI {
			return as
		}, a)

		xa, e := x.SafeResolve()
		if e != nil {
			t.Fatalf("Resolve failed %+v", e)
		}

		xa.Hello()

		b := Make[BI](func(as BI) BI {
			return as
		}, pa)

		xb, e := b.SafeResolve()
		if e != nil {
			t.Fatalf("Resolve failed %+v", e)
		}

		xb.Goodbye()

	})

	t.Run("can use init and reset to init and reset", func(t *testing.T) {
		i1 := Provide(func() int {
			return 1
		})

		i2 := Make[int](func(i int) int {
			return i + 2
		}, i1)

		i3 := Make[int](func(i int) int {
			return i + 3
		}, i2)

		var r int
		r = i3.Resolve()
		if r != 6 {
			t.FailNow()
		}

		i1.Reset()
		i1.Init(-5)
		i2.Reset()
		i3.Reset()

		r = i3.Resolve()
		if r != 0 {
			t.FailNow()
		}
	})

	t.Run("error should be treated well", func(t *testing.T) {
		ae := ProvideWithError(func() (int, error) {
			return 0, fmt.Errorf("error_0")
		})

		_, e := ae.SafeResolve()
		if e == nil {
			t.FailNow()
		}

		me := Make[int](func() (int, error) {
			return 0, fmt.Errorf("error 2")
		})

		_, e = me.SafeResolve()
		if e == nil {
			t.FailNow()
		}

		ce := Make[int](func(i int) (int, error) {
			return 0, fmt.Errorf("error 3")
		}, ae)

		_, e = ce.SafeResolve()
		if e == nil || !strings.Contains(e.Error(), "error_0") {
			t.FailNow()
		}
	})
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
