package submodule_test

import "github.com/submodule-org/submodule.go"

func Example_Make() {
	type counter struct {
		Start int
	}
	type Counter interface {
	}

	startPoint := submodule.Provide[int](func() int {
		return 0
	})

	submodule.Make[Counter](func(start int) Counter {
		return &counter{start}
	}, startPoint)
}
