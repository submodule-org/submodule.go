# Submodule

Function is not a module, because function cannot retain state
Module is a little bit too much, and module is less reusable

Submodule fits right in the middle, a little bit more than a function, a little bit less than a module

# Design prinicple

- Fail fast
- Function is the unit, answer what function wants. If function is the target unit, function will less likely to have unnecessary dependency
- Good DX must come with good Testability. Good DX and hard to test is an indicator of an unhealthy trade-offs