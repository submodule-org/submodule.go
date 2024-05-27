package mhttp

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/submodule-org/submodule.go"
	"github.com/submodule-org/submodule.go/meta/mconfig"
)

type rawServerConfig struct {
	Addr              string
	KeepAlive         bool
	ReadTimeout       string
	ReadHeaderTimeout string
	WriteTimeout      string
	IdleTimeout       string
	MaxHeaderBytes    string
}

var rawConfigMod = mconfig.CreateConfigWithPath("http", &rawServerConfig{
	Addr:              ":8080",
	KeepAlive:         true,
	ReadTimeout:       "5s",
	ReadHeaderTimeout: "5s",
	WriteTimeout:      "5s",
	IdleTimeout:       "60s",
	MaxHeaderBytes:    "1M",
})

type serverConfig struct {
	Addr              string
	KeepAlive         bool
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    uint64
}

var configMod = submodule.Make[*serverConfig](func(config *rawServerConfig) (s *serverConfig, e error) {
	s = &serverConfig{}

	s.ReadTimeout, e = time.ParseDuration(config.ReadTimeout)
	if e != nil {
		return nil, e
	}

	s.ReadHeaderTimeout, e = time.ParseDuration(config.ReadHeaderTimeout)
	if e != nil {
		return nil, e
	}

	s.WriteTimeout, e = time.ParseDuration(config.WriteTimeout)
	if e != nil {
		return nil, e
	}

	s.IdleTimeout, e = time.ParseDuration(config.IdleTimeout)
	if e != nil {
		return nil, e
	}

	s.MaxHeaderBytes, e = humanize.ParseBytes(config.MaxHeaderBytes)
	if e != nil {
		return nil, e
	}

	return s, nil
}, rawConfigMod)

var ServerMod = submodule.Make[*http.Server](func(self submodule.Self, config *serverConfig) *http.Server {
	muxes := submodule.Find([]IntegrateWithHttpServer{}, self.Scope)
	fmt.Printf("found %d handlers\n", len(muxes))
	fmt.Printf("config %+vv\n", config)

	rootMux := http.NewServeMux()

	for _, m := range muxes {
		m.AdaptToHTTPHandler(rootMux)
	}

	s := &http.Server{Handler: rootMux}

	s.SetKeepAlivesEnabled(config.KeepAlive)
	s.Addr = config.Addr
	s.ReadTimeout = config.ReadTimeout
	s.WriteTimeout = config.WriteTimeout

	return s
}, configMod)

func Start() error {
	server, e := ServerMod.SafeResolve()
	if e != nil {
		return e
	}

	return server.ListenAndServe()
}

func StartIn(scope submodule.Scope) error {
	fmt.Printf("starting server")
	server, e := ServerMod.SafeResolveWith(scope)
	if e != nil {
		return e
	}

	return server.ListenAndServe()
}

func Stop() error {
	fmt.Printf("stopping server")
	server, e := ServerMod.SafeResolve()
	if e != nil {
		return e
	}

	return server.Close()
}

func StopIn(scope submodule.Scope) error {
	server, e := ServerMod.SafeResolveWith(scope)
	if e != nil {
		return e
	}

	return server.Close()
}
