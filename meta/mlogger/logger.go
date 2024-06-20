package mlogger

import (
	"log/slog"

	"github.com/submodule-org/submodule.go/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
)

type LoggerConfig = zap.Config

func defaultConfig() LoggerConfig {
	return zap.NewDevelopmentConfig()
}

var defaultConfigMod = submodule.Value(defaultConfig())

var zapMod = submodule.MakeModifiable[*zap.Logger](func(config LoggerConfig) (*zap.Logger, error) {
	return config.Build()
}, defaultConfigMod)

func Alter(m func(config *LoggerConfig)) {
	c := defaultConfig()
	m(&c)
	zapMod.Append(submodule.Value(c))
}

func Reset() {
	zapMod.Reset()
}

func CreateLogger(name string, attrs ...any) submodule.Submodule[*slog.Logger] {
	return submodule.Make[*slog.Logger](func(logger *zap.Logger) *slog.Logger {
		l := slog.New(zapslog.NewHandler(logger.Core(), &zapslog.HandlerOptions{
			AddSource:  true,
			LoggerName: name,
		}))
		return l.With(attrs...)
	}, zapMod)
}
