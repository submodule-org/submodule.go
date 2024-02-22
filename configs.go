package submodule

type Mode int8

const (
	Prototype Mode = iota
	Singleton
)

func buildConfig(configs ...ConfigFn) Config {
	c := defaultConfig
	for _, cm := range configs {
		c = cm(c)
	}
	return c
}

type Config struct{ mode Mode }

var defaultConfig = Config{
	mode: Singleton,
}

type ConfigFn = func(config Config) Config

var SetPrototype = func(config Config) Config {
	config.mode = Prototype
	return config
}

var SetSingleton = func(config Config) Config {
	config.mode = Singleton
	return config
}
