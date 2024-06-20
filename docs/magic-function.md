# Magic function

Submodule comes with its most important operator called `Make`. And `Make` requires an `interface{}` as its first argument.

The main reason for this is because Golang is a statically typed language, and it's not possible to have a function that can accept any type of input.

As such, it'll require a document to express what a magic function is capable of.

## can be a no dependency function
```go
var strVal = submodule.Make[string](func() string {
  return "hello"
})
```

## can reports error as well
```go
var strVal = submodule.Make[string](func() (string, error) {
  return "", fmt.Errorf("error")
})
```

## can declare dependencies

```go

var intValue = submodule.Make[int](func() int {
  return 100
})

// ❌ this code will cause panic, because the `int` depedenncy is not provided
var strValue = submodule.Make[string](func(i int) string {
  return fmt.Sprintf("%d", i)
})

// ✅ this code will work
var strValue = submodule.Make[string](func(i int) string {
  return fmt.Sprintf("%d", i)
}, intValue)
```

## can have multiple dependencies
```go
var intValue = submodule.Make[int](func() int {
  return 100
})

var stringValue = submodule.Make[string](func(i int) string {
  return fmt.Sprintf("%d", i)
}, intValue)

var anotherStrValue = submodule.Make[string](func(i int, s string) string {
  return fmt.Sprintf("%d %s", i, s)
}, stringValue, intValue)
```

## dependencies order matters, the first one will be resolved and used (in case of same type)
```go
var intValue = submodule.Make[int](func() int {
  return 100
})

var stringValue = submodule.Make[string](func(i int) string {
  return fmt.Sprintf("%d", i)
}, intValue)

var anotherStrValue = submodule.Make[string](func(i int, s string) string {
  return fmt.Sprintf("%d %s", i, s)
}, stringValue, intValue)

var finalStrValue = submodule.Make[string](func(s string) string {
  return fmt.Sprintf("final %s", s)
  // anotherStrValue will be used
}, anotherStrValue, stringValue)

```

## magic parameter using `submodule.In`

A list of parameter can be too long, that'll make the code hard to read. To solve this, `submodule.In` can be used to declare a list of parameters
```go
var anotherStrValue = submodule.Make[string](func(i int, s string) string {
  return fmt.Sprintf("%d %s", i, s)
}, stringValue, intValue)

// is the same as
var anotherStrValue = submodule.Make[string](func(p struct {
  submodule.In
  // must be exported
  I int
  // must be exported
  S string
}) string {
  return fmt.Sprintf("%d %s", i, s)
}, stringValue, intValue)
```

## magic resolve

It's not directly to `Make`, but there's another operator called `Resolve`, it'll do exactly the same as magic parameter, but with the input struct

```go
type Service struct {
  // must be exported
  Config Config
  // must be exported
  Logger Logger
}

func NewService(config Config, logger Logger) *Service {
  return &Service{
    Config: config,
    Logger: logger,
  }
}

var service = submodule.Resolve(&Service{}, ConfigMod, LoggerMod)
// is the same as
var service = submodule.Make[*Service](NewService, ConfigMod, LoggerMod)
```