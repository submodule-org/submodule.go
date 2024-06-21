# Anti patterns and best practices

There are certain patterns that may ruin the experience of using submodule. There are some consequences

- defeat the purpose of using submodule, using normal way of coding golang is better
- make it even harder to test than normal golang code

## Singleton
Submodule only operates in Singleton mode against scope. Don't reimplement singleton

## Resolve and SafeResolve
Just don't use `Resolve` unless you know what you are doing. Resolve will just panic on error

## SafeResolve
Never use SafeResolve or Resolve in constructor function. You'll always resolve against the global scope

```go
// ❌ just don't do this
func NewService() *Service {
  return &Service{
    Config: ConfigMod.Resolve(),
    Logger: LoggerMod.Resolve(),
  }
}

// ❌ just don't do this
func NewService() (*Service, error) {
  config, err := ConfigMod.SafeResolve()
  if err != nil {
    return nil, err
  }

  logger, err := LoggerMod.SafeResolve()
  if err != nil {
    return nil, err
  }

  return &Service{
    Config: config,
    Logger: logger,
  }
}

// ✅ do this
func NewService(logger Logger, config Config) (*Service, error) {
  return &Service{
    Config: config,
    Logger: logger,
  }
}

// errors will be handled by submodule
var service = submodule.Make[*Service](NewService, ConfigMod, LoggerMod)
```

## When to use SafeResolve?

- in main function
```go
func main() {
  svc, e := service.SafeResolve()
  if e != nil {
    panic(e)
  } //...
}
```
- in test
```go
// ...
const scope = submodule.CreateScope()
svc, e := service.SafeResolveWith(scope)
// test
```
- in few rare cases, within framework handler, but that means we will not be able to test it