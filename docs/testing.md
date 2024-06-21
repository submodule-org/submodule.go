# how to test submodule

Submodule is designed to be testable.

## Always test against specific scope
```go
func TestService(t *testing.T) {
  scope := submodule.CreateScope()
  svc, e := service.SafeResolveWith(scope)
  require.Nil(t, e)
 
  // test svc functions
}
```

## Emulates different situations via dependencies
```go
type Config struct { value string }
var configMod = submodule.Make[Config](func() Config {
  return Config{ value: "hello" }
})

type DB struct { config: Config }
var dbMod = //

type Svc struct { db: DB, config: Config }

func TestService(t *testing.T) {
  scope := submodule.CreateScope()

  // emulate Config to be specific value to see if the service will work
  configMod.ResolveToWith(scope, Config{ value: "world" })

  // can also replace dbMod with something else so can mock its operations
  dbMod.ResolveToWith(scope, DB{ /* different implementation */})

  // the whole dependency graph will be resolved automatically
  svc, e := service.SafeResolveWith(scope)
  require.Nil(t, e)

  // test svc functions
```