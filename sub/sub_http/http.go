package sub_http

import (
	"net/http"

	"github.com/submodule-org/submodule.go"
)

type Handler struct {
	Path    string
	Handler http.Handler
}

type CanHandleHTTP interface {
	GetHTTPHandler() Handler
}

var ServerMod = submodule.Make[http.Server](func(self submodule.Self) http.Server {
	muxes := submodule.Find([]CanHandleHTTP{}, self.Scope)

	rootMux := http.NewServeMux()

	for _, m := range muxes {
		mux := m.GetHTTPHandler()
		rootMux.Handle(mux.Path, mux.Handler)
	}

	return http.Server{
		Addr:    ":8080",
		Handler: rootMux,
	}
})

func Start() error {
	server, e := ServerMod.SafeResolve()
	if e != nil {
		return e
	}

	return server.ListenAndServe()
}
