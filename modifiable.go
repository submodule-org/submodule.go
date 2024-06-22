package submodule

import "reflect"

// ModifiableSubmodule is a submodule that can be modified.
// Overriding will be done by using the Append method mechanism, the default value will always be resolved last
// That'll help the submodule to be way easier to reconfigure (for example, sharing loggers) without missing the default settings
type ModifiableSubmodule[T any] interface {
	Submodule[T]
	Append(submodule ...Retrievable)
	Reset()
}

type modifiableSubmodule[T any] struct {
	submodule Submodule[T]
	modifiers []Retrievable
}

func (m *modifiableSubmodule[T]) Substitute(other Submodule[T]) {
	m.submodule.Substitute(other)
}

// Resolve implements ModifiableSubmodule.
func (m *modifiableSubmodule[T]) Resolve() T {
	return m.submodule.Resolve()
}

// ResolveTo implements ModifiableSubmodule.
func (m *modifiableSubmodule[T]) ResolveTo(t T) {
	m.submodule.ResolveTo(t)
}

// ResolveToWith implements ModifiableSubmodule.
func (m *modifiableSubmodule[T]) ResolveToWith(s Scope, t T) {
	m.submodule.ResolveToWith(s, t)
}

// ResolveWith implements ModifiableSubmodule.
func (m *modifiableSubmodule[T]) ResolveWith(s Scope) T {
	return m.submodule.ResolveWith(s)
}

// SafeResolve implements ModifiableSubmodule.
func (m *modifiableSubmodule[T]) SafeResolve() (T, error) {
	return m.submodule.SafeResolve()
}

// SafeResolveWith implements ModifiableSubmodule.
func (m *modifiableSubmodule[T]) SafeResolveWith(s Scope) (T, error) {
	return m.submodule.SafeResolveWith(s)
}

// canResolve implements ModifiableSubmodule.
func (m *modifiableSubmodule[T]) canResolve(t reflect.Type) bool {
	return m.submodule.canResolve(t)
}

// retrieve implements ModifiableSubmodule.
func (m *modifiableSubmodule[T]) retrieve(s Scope) (any, error) {
	return m.submodule.retrieve(s)
}

func (m *modifiableSubmodule[T]) Append(submodule ...Retrievable) {
	if len(submodule) == 0 {
		return
	}
	m.modifiers = append(m.modifiers, submodule...)
}

func (m *modifiableSubmodule[T]) Reset() {
	m.modifiers = []Retrievable{}
}

// Modifiable constructor. Parameters are the same as Make
func MakeModifiable[T any](fn any, dependencies ...Retrievable) ModifiableSubmodule[T] {
	xs := &modifiableSubmodule[T]{
		modifiers: []Retrievable{},
	}

	xs.submodule = Make[T](func(self Self) (t T, e error) {
		var modifiedDependencies []Retrievable
		modifiedDependencies = append(modifiedDependencies, xs.modifiers...)
		modifiedDependencies = append(modifiedDependencies, dependencies...)

		return Make[T](fn, modifiedDependencies...).SafeResolveWith(self.Scope)
	}, dependencies...)

	return xs
}
