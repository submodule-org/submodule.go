---
name: submodule-go
description: Use when writing Go code that uses the submodule.go DI framework (github.com/submodule-org/submodule.go/v2). Triggers on: imports of "submodule.go/v2", usage of submodule.Make, submodule.Resolve, submodule.Value, submodule.Group, submodule.MakeModifiable, submodule.CreateScope, any Submodule[T] type, or when the user asks about dependency injection in Go with this library. Also triggers when writing tests for code that uses submodule DI, or when the user mentions "submodule" in a Go context. Use this skill even when the task seems simple — incorrect DI registration silently breaks testability.
---

# submodule.go DI Framework

`submodule.go` is a lightweight, type-safe Dependency Injection framework for Go built on function composition and lazy evaluation. Dependencies are wrapped in factory functions and only initialized when resolved. Each value is singleton within its scope.

Import: `github.com/submodule-org/submodule.go/v2`

## Core Concept

A `Submodule[T]` is a container holding a factory function and its dependency metadata. It does not hold the result — values are resolved against a `Scope` (a cache). This design enables:
- Lazy initialization: only what's needed gets created
- Testability: swap any dependency via scope isolation
- Explicit wiring: dependencies are declared, not hidden

## Registration API

### `Make[T](fn, ...dependencies)` — The primary registration function

The factory function `fn` is typed as `any` because Go can't express "any function signature" statically. The rules for what `fn` can be:

**Return types** (exactly one of):
- `func(...) T` — returns the value
- `func(...) (T, error)` — returns value + error

**Parameter types** (zero or more, in any combination):
- Concrete types that match a dependency's output type
- `struct { submodule.In; Field1 Type1; Field2 Type2 }` — groups parameters into a struct (all fields must be exported)
- `submodule.Self` — grants access to the current `Scope` and `Dependencies` at resolution time
- Variadic `...T` — the last parameter can be variadic

**Dependencies must be explicitly listed.** If a parameter type doesn't match any listed dependency, registration panics at init time (fail-fast).

```go
// No dependencies
var ConfigMod = submodule.Make[Config](func() Config {
    return Config{Host: "localhost", Port: 8080}
})

// With dependencies — each dependency is listed after the factory
var LoggerMod = submodule.Make[Logger](func(config Config) Logger {
    return &logger{LogLevel: config.LogLevel}
}, ConfigMod)

// With error return
var DbMod = submodule.Make[*DB](func(config Config) (*DB, error) {
    return sql.Open("postgres", config.DSN)
}, ConfigMod)

// Using submodule.In for many parameters
var ServiceMod = submodule.Make[*Service](func(p struct {
    submodule.In
    Config Config
    Logger Logger
    DB     *DB
}) *Service {
    return &Service{Config: p.Config, Logger: p.Logger, DB: p.DB}
}, ConfigMod, LoggerMod, DbMod)

// Using submodule.Self for scope access
var WorkerMod = submodule.Make[*Worker](func(self submodule.Self) *Worker {
    self.Scope.AppendMiddleware(submodule.WithScopeEnd(func() error {
        return cleanup()
    }))
    return &Worker{}
})
```

**Dependency order matters when multiple dependencies provide the same type.** The first matching dependency is used:
```go
var a = submodule.Make[string](func() string { return "a" })
var b = submodule.Make[string](func() string { return "b" })

// s resolves to "a" because `a` is listed first
var s = submodule.Make[string](func(v string) string { return v }, a, b)
```

### `Resolve[T](structInstance, ...dependencies)` — Struct field injection

Resolves exported struct fields against dependencies. This is syntactic sugar over `Make` with a struct parameter.

```go
type Server struct {
    Config Config   // exported fields only
    Logger Logger
}

// Both are equivalent:
var ServerMod = submodule.Resolve(&Server{}, ConfigMod, LoggerMod)
var ServerMod = submodule.Make[*Server](func(c Config, l Logger) *Server {
    return &Server{Config: c, Logger: l}
}, ConfigMod, LoggerMod)
```

Only exported fields are resolved. Unexported fields cause an error.

### `Value[T](value)` — Register a static value

No factory, no dependencies. Good for configs, constants, and test overrides.

```go
var DefaultTimeout = submodule.Value(30 * time.Second)
```

### `Group[T](...submodules)` — Aggregate into a slice

Collects multiple submodules of the same type into `[]T`.

```go
var AllHandlers = submodule.Group[Handler](UserHandler, OrderHandler, AdminHandler)
```

### `MakeModifiable[T](fn, ...dependencies)` — Overridable dependencies

Like `Make`, but dependencies can be appended/overridden later without creating a new submodule. Appended dependencies take priority over original ones.

```go
var LoggerMod = submodule.MakeModifiable[*slog.Logger](func(config LoggerConfig) (*slog.Logger, error) {
    return config.Build()
}, defaultConfigMod)

// Override the config dependency
LoggerMod.Append(submodule.Value(productionConfig))

// Reset overrides
LoggerMod.Reset()
```

## Resolving Values

```go
// Against global scope (singleton for app lifetime)
value := MyMod.Resolve()            // panics on error
value, err := MyMod.SafeResolve()   // returns error

// Against a custom scope
scope := submodule.CreateScope()
value := MyMod.ResolveWith(scope)
value, err := MyMod.SafeResolveWith(scope)

// Force a specific value (bypass factory)
MyMod.ResolveTo(overrideValue)                  // global scope
MyMod.ResolveToWith(scope, overrideValue)        // custom scope
```

## Scopes

A scope is a value cache. Each submodule resolves at most once per scope (singleton within scope).

```go
// Global scope — shared across the app
globalScope := submodule.GetStore()

// Isolated scope — fresh cache, no shared state
scope := submodule.CreateScope()

// Inherited scope — reads from global cache first
scope := submodule.CreateScope(submodule.Inherit(true))

// Child scope — reads from parent cache first
child := submodule.CreateScope(submodule.WithParent(parentScope))

// Scope with middleware
scope := submodule.CreateScope(submodule.WithMiddlewares(myMiddleware))
```

### Disposing scopes

Disposal triggers `WithScopeEnd` and `WithContextScopeEnd` middleware in reverse order, then clears all cached values.

```go
err := scope.Dispose()
err := scope.DisposeWithContext(ctx)

// Global scope
err := submodule.DisposeGlobalScope()
err := submodule.DisposeGlobalScopeWithContext(ctx)
```

## Middleware

Middleware decorates resolved values or hooks into scope disposal.

```go
// Decorate a resolved value of type T
submodule.WithScopeResolve(func(db *DB) *DB {
    return wrapWithMetrics(db)
})

// Run cleanup when scope disposes
submodule.WithScopeEnd(func() error {
    return db.Close()
})

// Run cleanup with context
submodule.WithContextScopeEnd(func(ctx context.Context) error {
    return db.CloseWithContext(ctx)
})

// Add middleware from within a factory via Self
var DbMod = submodule.Make[*DB](func(self submodule.Self) *DB {
    db := connectDB()
    self.Scope.AppendMiddleware(submodule.WithContextScopeEnd(func(ctx context.Context) error {
        return db.Close()
    }))
    return db
})
```

## Testing — The Critical Part

Testing is the primary reason submodule.go exists. Follow these patterns strictly.

### Always use an isolated scope

Never resolve against the global scope in tests — it leaks state between test cases.

```go
func TestService(t *testing.T) {
    scope := submodule.CreateScope()
    svc, err := ServiceMod.SafeResolveWith(scope)
    require.NoError(t, err)

    // test svc...
}
```

### Mock via ResolveToWith

Override any dependency in the graph by forcing a value in the test scope. The entire dependency chain re-resolves automatically with your override.

```go
func TestServiceWithMockDB(t *testing.T) {
    scope := submodule.CreateScope()

    // Override the DB dependency with a mock
    DbMod.ResolveToWith(scope, &mockDB{})

    // Service will receive the mock DB
    svc, err := ServiceMod.SafeResolveWith(scope)
    require.NoError(t, err)

    // test svc...
}
```

### Mock via Substitute

Replaces a submodule's factory entirely. Use when you need a different factory, not just a different value.

```go
var mockServiceMod = submodule.Make[Service](func() Service {
    return &MockService{}
})

func TestHigherLevel(t *testing.T) {
    ServiceMod.Substitute(mockServiceMod)
    // all dependents now use MockService
}
```

### Simulate errors

```go
func TestServiceHandlesDBError(t *testing.T) {
    scope := submodule.CreateScope()
    scope.InitError(DbMod, fmt.Errorf("connection refused"))

    _, err := ServiceMod.SafeResolveWith(scope)
    require.Error(t, err)
    require.Contains(t, err.Error(), "connection refused")
}
```

## Anti-Patterns — NEVER Do These

These patterns silently destroy testability. The code compiles and runs, but tests become impossible to isolate.

### NEVER call Resolve/SafeResolve inside a factory function

This resolves against the global scope, bypassing the dependency graph. The test scope cannot intercept these calls.

```go
// WRONG — resolves against global scope, untestable
func NewService() *Service {
    return &Service{
        Config: ConfigMod.Resolve(),
        Logger: LoggerMod.Resolve(),
    }
}

// WRONG — same problem with SafeResolve
func NewService() (*Service, error) {
    config, err := ConfigMod.SafeResolve()
    if err != nil { return nil, err }
    return &Service{Config: config}, nil
}

// CORRECT — accept dependencies as parameters
func NewService(config Config, logger Logger) *Service {
    return &Service{Config: config, Logger: logger}
}
var ServiceMod = submodule.Make[*Service](NewService, ConfigMod, LoggerMod)
```

### NEVER re-implement singleton

Submodule already guarantees singleton within a scope. Adding your own sync.Once or package-level var defeats scope isolation.

### Use SafeResolve over Resolve

`Resolve()` panics on error. Only use it in `main()` or when you explicitly want a panic. In tests, always use `SafeResolveWith(scope)`.

### Where SafeResolve/Resolve IS appropriate

- In `main()` to bootstrap the app
- In tests with `SafeResolveWith(scope)`
- Rarely, in framework handlers (but this makes them untestable)

## Module Organization Pattern

Declare submodule variables at package level. Group related dependencies in the same file.

```go
// db/module.go
package db

var ConfigMod = submodule.Make[Config](loadConfig)
var ConnMod = submodule.Make[*sql.DB](func(c Config) (*sql.DB, error) {
    return sql.Open("postgres", c.DSN)
}, ConfigMod)

// handlers/module.go
package handlers

var UserHandlerMod = submodule.Resolve(&UserHandler{}, db.ConnMod, logger.Mod)
```

## Interface Pattern

Expose dependencies as interfaces to enable clean mocking:

```go
type Repository interface {
    FindByID(id string) (*User, error)
}

// Register the concrete implementation behind the interface
var RepoMod = submodule.Make[Repository](func(db *sql.DB) Repository {
    return &pgRepo{db: db}
}, DbMod)
```

## Quick Reference

| Function | Purpose | Example |
|---|---|---|
| `Make[T](fn, deps...)` | Register factory with dependencies | `Make[Logger](NewLogger, ConfigMod)` |
| `Resolve[T](struct, deps...)` | Auto-inject struct fields | `Resolve(&Server{}, ConfigMod)` |
| `Value[T](v)` | Register static value | `Value(Config{Port: 8080})` |
| `Group[T](subs...)` | Collect into `[]T` | `Group[Handler](H1, H2)` |
| `MakeModifiable[T](fn, deps...)` | Overridable dependencies | `MakeModifiable[*Logger](NewLogger, CfgMod)` |
| `CreateScope(opts...)` | New isolated scope | `CreateScope()` |
| `SafeResolveWith(scope)` | Resolve in scope (safe) | `Mod.SafeResolveWith(scope)` |
| `ResolveToWith(scope, v)` | Override value in scope | `Mod.ResolveToWith(scope, mock)` |
| `Substitute(other)` | Replace factory entirely | `Mod.Substitute(mockMod)` |
| `scope.InitError(mod, err)` | Inject error in scope | `scope.InitError(DbMod, err)` |
| `scope.Dispose()` | Cleanup scope | `scope.Dispose()` |
