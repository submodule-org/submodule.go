package subs

type ConfigSetter[C, V any] interface {
	WithManually(C) *V
	WithDefault() *V
}
