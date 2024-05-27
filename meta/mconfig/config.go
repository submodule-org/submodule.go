package mconfig

import (
	"github.com/spf13/viper"
	"github.com/submodule-org/submodule.go"
)

type config struct {
	configName  string
	configType  string
	configPaths []string
	envPrefix   string
}

var defaults = config{
	configName:  "app",
	configType:  "env",
	configPaths: []string{"."},
	envPrefix:   "",
}

var inuse = config{
	configName: defaults.configName,
	configType: defaults.configType,

	configPaths: defaults.configPaths,
	envPrefix:   defaults.envPrefix,
}

func SetConfigName(n string) {
	inuse.configName = n
}

func SetConfigType(t string) {
	inuse.configType = t
}

func SetEnvPrefix(p string) {
	inuse.envPrefix = p
}

func AppendConfigPath(p string) {
	inuse.configPaths = append(inuse.configPaths, p)
}

func SetDefaults() {
	inuse = defaults
}

var viperMod = submodule.Make[*viper.Viper](func() (*viper.Viper, error) {
	var e error
	v := viper.New()

	v.SetEnvPrefix(inuse.envPrefix)

	v.SetConfigName(inuse.configName)
	v.SetConfigType(inuse.configType)

	for _, p := range inuse.configPaths {
		v.AddConfigPath(p)
	}

	v.AutomaticEnv()

	if e := v.ReadInConfig(); e != nil {
		if _, ok := e.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
		} else {
			return nil, e
		}
	}

	return v, e
})

type ConfigLoader struct {
	viper *viper.Viper
}

func (c *ConfigLoader) Load(t any) error {
	return c.viper.Unmarshal(&t)
}

func (c *ConfigLoader) LoadPath(p string, t any) error {
	return c.viper.UnmarshalKey(p, &t)
}

var LoaderMod = submodule.Make[*ConfigLoader](func(viper *viper.Viper) *ConfigLoader {
	return &ConfigLoader{viper: viper}
}, viperMod)
