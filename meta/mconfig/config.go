package mconfig

import (
	"os"
	"strings"

	"github.com/spf13/viper"
	"github.com/submodule-org/submodule.go"
)

type ConfigModifier func(*viper.Viper)

type config struct {
	configName  string
	configType  string
	configPaths []string
	envPrefix   string
	modifiers   []ConfigModifier
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

func AppendDefault(k string, v any) {
	inuse.modifiers = append(inuse.modifiers, func(c *viper.Viper) {
		c.SetDefault(k, v)
	})
}

var viperMod = submodule.Make[*viper.Viper](func() (*viper.Viper, error) {
	var e error
	v := viper.New()

	replacer := strings.NewReplacer(`_`, `.`)

	v.SetEnvKeyReplacer(replacer)
	v.SetEnvPrefix(inuse.envPrefix)
	v.SetConfigName(inuse.configName)
	v.SetConfigType(inuse.configType)

	for _, m := range inuse.modifiers {
		m(v)
	}

	for _, p := range inuse.configPaths {
		v.AddConfigPath(p)
	}

	for _, s := range os.Environ() {
		a := strings.Split(s, "=")
		v.Set(replacer.Replace(a[0]), a[1])
	}

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

func (c *ConfigLoader) LoadWithDefault(t any) error {
	return c.viper.Unmarshal(&t)
}

func (c *ConfigLoader) LoadPath(p string, t any) error {
	return c.viper.UnmarshalKey(p, &t)
}

func (c *ConfigLoader) LoadPathWithDefault(p string, d any, t any) error {
	c.viper.SetDefault(p, d)

	return c.viper.UnmarshalKey(p, &t)
}

var LoaderMod = submodule.Make[*ConfigLoader](func(viper *viper.Viper) *ConfigLoader {
	return &ConfigLoader{viper: viper}
}, viperMod)

func CreateConfig[T any](def T) submodule.Submodule[T] {
	return CreateConfigWithPath("", def)
}

func CreateConfigWithPath[T any](path string, def T) submodule.Submodule[T] {
	AppendDefault(path, def)
	return submodule.Make[T](func(loader *ConfigLoader) T {
		loader.Load(def)
		return def
	}, LoaderMod)
}
