package submodule

// Deprecated: No longer supports Prototype or Singleton, use function instead
type Mode int8

// Deprecated: No longer supports Prototype or Singleton, use function instead
const (
	Prototype Mode = iota
	Singleton
)

// Deprecated: No longer supports Prototype or Singleton, use function instead
type Config struct{ mode Mode }

// Deprecated: No longer supports Prototype or Singleton, use function instead
type ConfigFn = func(config Config) Config

// Deprecated: No longer supports Prototype or Singleton, use function instead
var SetPrototype = func(config Config) Config {
	config.mode = Prototype
	return config
}

// Deprecated: No longer supports Prototype or Singleton, use function instead
var SetSingleton = func(config Config) Config {
	config.mode = Singleton
	return config
}
