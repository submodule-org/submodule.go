# Scope

Scope is where values are stored. A scope guarantees a submodule to be resolved only once, and will be disposed when the scope is disposed.

In short, a value of a submodule will be singleton within the scope

## Global scope
```
var scope = submodule.GetStore()
```

## Create a scope

```go
var scope = submodule.CreateScope()
```

## To actualize a submodule
```go
var value = valueMod.Resolve()
// is the same as
var value = valueMod.ResolveWith(GetStore())
```

## Singleton
```go
var value = valueMod.Resolve()
value = valueMod.Resolve()
value = valueMod.Resolve()
value = valueMod.Resolve()
value = valueMod.Resolve()
// all those value are just the same, cached by the global scope

GetStore().Dispose()

// now all those value will be disposed
value = valueMod.Resolve()
// this is a new value
value = valueMod.Resolve()
value = valueMod.Resolve()
// same new value
```
