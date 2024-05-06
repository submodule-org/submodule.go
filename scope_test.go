package submodule_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/submodule-org/submodule.go"
)

func TestScope(t *testing.T) {
	var seed int
	intValue := submodule.Make[int](func(self submodule.Self) int {
		self.Scope.AppendMiddleware(submodule.WithScopeEnd(func() error {
			seed = seed + 2
			return nil
		}))

		return seed
	})

	replaceInt := submodule.WithScopeResolve(func(i int) int {
		fmt.Println("replacing 0 with 4")
		return i + 1
	})

	onEnd := submodule.WithScopeEnd(func() error {
		fmt.Println("scope is ended")
		return nil
	})

	scope := submodule.CreateScope(
		submodule.WithMiddlewares(replaceInt, onEnd),
	)

	v := intValue.ResolveWith(scope)
	assert.Equal(t, 1, v)

	e := scope.Dispose()
	assert.Nil(t, e)

	assert.Equal(t, 2, seed)
}
