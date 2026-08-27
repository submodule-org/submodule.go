package submodule_test

import (
	"testing"

	"github.com/submodule-org/submodule.go/v2"
)

// Resolution is cached per scope, so each iteration needs a fresh scope.
// The delta between the no-middleware and with-middleware benchmarks isolates
// the cost of the scope-resolve middleware path.
func benchmarkResolveWithMiddlewares(b *testing.B, mws ...submodule.Middleware) {
	mod := submodule.Make[int](func() int { return 1 })

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		scope := submodule.CreateScope(submodule.WithMiddlewares(mws...))
		if v := mod.ResolveWith(scope); v == 0 {
			b.Fatal("unexpected zero")
		}
	}
}

func BenchmarkResolve_NoMiddleware(b *testing.B) {
	benchmarkResolveWithMiddlewares(b)
}

func BenchmarkResolve_FourMiddlewares(b *testing.B) {
	inc := func(i int) int { return i + 1 }
	benchmarkResolveWithMiddlewares(b,
		submodule.WithScopeResolve(inc),
		submodule.WithScopeResolve(inc),
		submodule.WithScopeResolve(inc),
		submodule.WithScopeResolve(inc),
	)
}
