package submodule_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/submodule-org/submodule.go/v2"
)

func TestScope(t *testing.T) {
	var seed int
	var isDisposeGlobal bool
	var isDisposeLocal bool
	intValue := submodule.Make[int](func(self submodule.Self) int {
		self.Scope.AppendMiddleware(
			submodule.WithScopeEnd(func() error {
				fmt.Println("add 2 at the end of the day")
				seed = seed + 2
				isDisposeLocal = true
				return nil
			}),
			submodule.WithScopeResolve(func(i any) any {
				fmt.Println("catch everything")
				return i
			}),
		)
		return seed
	})

	replaceInt := submodule.WithScopeResolve(func(i int) int {
		fmt.Println("replacing 0 with 4")
		return i + 1
	})

	onEnd := submodule.WithScopeEnd(func() error {
		fmt.Println("scope is ended")
		isDisposeGlobal = true
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
	assert.True(t, isDisposeGlobal)
	assert.True(t, isDisposeLocal)
}

func Test_Scope_With_Context(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var isDispose bool
	sub := submodule.Make[int](func(self submodule.Self) int {
		self.Scope.AppendMiddleware(
			submodule.WithContextScopeEnd(func(ctx context.Context) error {
				<-ctx.Done()
				isDispose = true
				return nil
			}),
		)
		return 0
	})
	scope := submodule.CreateScope()
	_ = sub.ResolveWith(scope)
	go func() {
		cancel()
	}()
	err := scope.DisposeWithContext(ctx)
	assert.Nil(t, err)
	assert.True(t, isDispose)
}
