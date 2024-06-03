package mhttp

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/submodule-org/submodule.go"
	"github.com/submodule-org/submodule.go/meta/mlogger"
)

type ServerConfig struct {
	Addr              string
	KeepAlive         bool
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    uint64
}

func defaultServerConfig() ServerConfig {
	return ServerConfig{
		Addr:              ":8080",
		KeepAlive:         true,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       10 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
}

var defaultServerConfigMod = submodule.Value(defaultServerConfig())
var defaultHttpLogger = mlogger.CreateLogger("http")

func AlterConfig(c func(*ServerConfig)) {
	mc := defaultServerConfig()
	c(&mc)
	Server.Append(submodule.Value(mc))
}

func Reset() {
	Server.Reset()
}

var Server = submodule.MakeModifiable[*http.Server](func(self submodule.Self, config ServerConfig, logger *slog.Logger) *http.Server {
	muxes := submodule.Find([]IntegrateWithHttpServer{}, self.Scope)
	logger.Debug("server is running with", "config", config)
	logger.Debug("found_routes %v", "muxes", muxes)

	rootMux := http.NewServeMux()

	for _, m := range muxes {
		m.AdaptToHTTPHandler(rootMux)
	}

	s := &http.Server{
		Handler: rootMux,
		Addr:    config.Addr,
	}

	s.SetKeepAlivesEnabled(config.KeepAlive)
	s.ReadTimeout = config.ReadTimeout
	s.WriteTimeout = config.WriteTimeout

	return s
}, defaultServerConfigMod, defaultHttpLogger)
