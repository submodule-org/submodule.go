package submodule

import (
	"fmt"
	"reflect"
	"testing"
)

func TestModuleFunction(t *testing.T) {
	t.Cleanup(func() {
		fmt.Println("Cleaning up")
		Show()
		Clean()
		Show()
	})

	t.Run("test resolve", func(t *testing.T) {

		Provide(func() (string, error) {
			return "test", nil
		})

		x := reflect.TypeOf("test")

		v, e := Resolve(x)
		if e != nil {
			t.Fatalf("Resolve failed %+v", e)
		}

		if v != "test" {
			t.Fatal("Resolve failed")
		}
	})

	t.Run("Resolve by fields", func(t *testing.T) {
		Provide(func() (string, error) {
			return "test", nil
		})

		Provide(func() (int, error) {
			return 3, nil
		})

		type Test struct {
			A string
			B int
		}

		x := reflect.TypeOf(Test{})

		v, e := ResolveByFields(x)
		if e != nil {
			t.Fatalf("Resolve failed %+v", e)
		}

		fmt.Printf("Resolved %+v\n", v)

		if v.(Test).A != "test" {
			t.Fatal("Resolve failed")
		}

		if v.(Test).B != 3 {
			t.Fatal("Resolve failed")
		}
	})

	t.Run("derive", func(t *testing.T) {
		Provide(func() (string, error) {
			return "test", nil
		})

		Derive(func(p struct{ A string }) (int, error) {
			if p.A == "test" {
				return 3, nil
			}
			return 0, fmt.Errorf("Failed")
		})

		type Test struct {
			A string
			B int
		}

		x := reflect.TypeOf(Test{})

		v, e := ResolveByFields(x)
		if e != nil {
			t.Fatalf("Resolve failed %+v", e)
		}

		fmt.Printf("Resolved %+v\n", v)

		if v.(Test).A != "test" {
			t.Fatal("Resolve failed")
		}

		if v.(Test).B != 3 {
			t.Fatal("Resolve failed")
		}
	})

	t.Run("execute", func(t *testing.T) {
		Provide(func() (int, error) {
			return 45, nil
		})

		v, error := Execute(func(p struct{ X int }) (int, error) {
			return p.X, nil
		})

		if error != nil {
			t.Fatalf("Execute failed %+v", error)
		}

		if v != 45 {
			t.FailNow()
		}
	})

	t.Run("direct call", func(t *testing.T) {
		makeInt := Provide(func() (int, error) {
			return 45, nil
		})

		v1, e1 := makeInt()
		if e1 != nil {
			t.Fatalf("Execute failed %+v", e1)
		}

		if v1 != 45 {
			t.FailNow()
		}

		makeStringFromInt := Derive(func(p struct{ X int }) (string, error) {
			return fmt.Sprintf("%d", p.X), nil
		})

		v2, e2 := makeStringFromInt()
		if e2 != nil {
			t.Fatalf("Execute failed %+v", e2)
		}

		if v2 != "45" {
			t.FailNow()
		}

	})

	t.Run("test error", func(t *testing.T) {
		// Derive should be error as well
		makeStringFromInt := Derive(func(p struct{ X int }) (string, error) {
			return "", nil
		})

		v, e2 := makeStringFromInt()
		if e2 != nil {
			t.Fatalf("Must have error %s, but got %+v", e2, v)

			if e2.Error() != "failed to resolve int" {
				t.Fatalf("Error didn't carry over %s", v)
			}
		}

	})

}
