package mhttp

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/submodule-org/submodule.go"
	"github.com/submodule-org/submodule.go/meta/mconfig"
)

type serverConfig struct {
	ConfigPath string
	Addr       string

	DisableGeneralOptionsHandler bool
	TLSConfig                    *tls.Config
	ReadTimeout                  time.Duration

	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
}

var defaultServerConfig = serverConfig{
	ConfigPath:   "http",
	Addr:         ":8080",
	ReadTimeout:  5 * time.Second,
	WriteTimeout: 10 * time.Second,
}

var configInUse = defaultServerConfig

func UseDefault() {
	configInUse = defaultServerConfig
}

func SetAddr(addr string) {
	configInUse.Addr = addr
}

func SetConfigPath(path string) {
	configInUse.ConfigPath = path
}

type IntegrateWithHttpServer interface {
	AdaptToHTTPHandler(rootMux *http.ServeMux)
}

var configMod = submodule.Make[*serverConfig](func(loader *mconfig.ConfigLoader) (*serverConfig, error) {
	e := loader.LoadPath(configInUse.ConfigPath, &configInUse)

	if e != nil {
		return nil, e
	}

	return &configInUse, nil
}, mconfig.LoaderMod)

var ServerMod = submodule.Make[*http.Server](func(self submodule.Self, config *serverConfig) *http.Server {
	muxes := submodule.Find([]IntegrateWithHttpServer{}, self.Scope)
	fmt.Printf("found %d handlers", len(muxes))
	fmt.Printf("config %+v", config)

	rootMux := http.NewServeMux()

	for _, m := range muxes {
		m.AdaptToHTTPHandler(rootMux)
	}

	s := &http.Server{Handler: rootMux}

	s.SetKeepAlivesEnabled(true)
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
