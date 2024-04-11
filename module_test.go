package submodule

import (
	"fmt"
	"testing"
)

func ms() Submodule[string] {
	return Make[string](func() string {
		return "hello"
	})
}

func TestModuleFunction(t *testing.T) {

	t.Run("test module function", func(t *testing.T) {
		s, e := ms().Resolve()
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

		xb, e := b.Resolve()
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

		Make[Bi](func() Ai {
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

		Override(b, aa)

		xb, e := b.Resolve()
		if e != nil {
			t.Fatalf("Resolve failed %+v", e)
		}

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

		_, e := s.Resolve()
		if e != nil {
			t.Fatalf("Resolve failed %+v", e)
		}

		ni, _ := s.Resolve()

		if ni != 1 {
			fmt.Printf("%+v\n", ni)
			t.FailNow()
		}
	})

	t.Run("expose as interface", func(t *testing.T) {
		s := Make[As](func() Ai {
			return As{}
		})

		xs, e := s.Resolve()
		if e != nil {
			t.FailNow()
		}

		xs.Hello()
	})

	t.Run("annotate with tag", func(t *testing.T) {
		type A struct {
			Hello string
		}

		s := Provide(
			func() A { return A{} },
			WithTag(`name:"hello"`),
		)

		_, e := s.Resolve()
		if e != nil {
			fmt.Printf("caught error %+v\n", e)
			t.FailNow()
		}

		sm, _ := s.(*submodule[A])
		if sm.tagger.tag != `name:"hello"` {
			t.FailNow()
		}

	})

	t.Run("in can be resolved", func(t *testing.T) {
		type A struct {
			Hello string
		}

		type B struct {
			Prefix string
		}

		a := Provide(func() A {
			return A{
				Hello: "hello",
			}
		})

		b := Make[B](func(a A) B {
			return B{
				Prefix: a.Hello,
			}
		}, a)

		x := Make[string](func(p struct {
			In
			A A
			B B
		}) string {
			return p.A.Hello + p.B.Prefix
		}, a, b)

		xx, e := x.Resolve()
		if e != nil {
			t.Fatalf("Resolve failed %+v", e)
		}

		if xx != "hellohello" {
			t.FailNow()
		}
	})

	t.Run("group lookup", func(t *testing.T) {
		type A struct {
			Id string
		}

		a1 := Provide(func() A {
			return A{
				Id: "1",
			}
		})

		a2 := Provide(func() A {
			return A{
				Id: "2",
			}
		})

		c := Make[int](func(p struct {
			In
			As []A `group:"true"`
		}) int {
			return len(p.As)
		}, a1, a2)

		xc, e := c.Resolve()

		if e != nil {
			t.Fatalf("Resolve failed %+v", e)
		}

		if xc != 2 {
			t.FailNow()
		}
	})
}

func TestUtils(t *testing.T) {

	t.Run("extract function should work", func(t *testing.T) {
		type A struct{}
		type B struct{}

		a := Provide(func() A {
			return A{}
		})

		b := Provide(func() B {
			return B{}
		})

		ca := WithTag(`name:"hello"`)

		gettables, alters := extract([]alterable{a, b, ca})
		if len(gettables) != 2 {
			t.FailNow()
		}

		if len(alters) != 1 {
			t.FailNow()
		}
	})

}

type As struct{}
type Ai interface {
	Hello()
}

type Bi interface {
	Goodbye()
}

func (a As) Hello() {
	fmt.Println("hello")
}
