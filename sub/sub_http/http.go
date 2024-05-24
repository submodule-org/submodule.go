package sub_http

import (
	"net/http"

	"github.com/submodule-org/submodule.go"
)

type GetMux interface {
	GetMux() *http.ServeMux
	GetMuxPath() string
}

var ServerMod = submodule.Make[http.Server](func(self submodule.Self) http.Server {
	muxes := submodule.Find([]GetMux{}, self.Scope)

	rootMux := http.NewServeMux()

	for _, m := range muxes {
		rootMux.Handle(m.GetMuxPath(), m.GetMux())
	}

	return http.Server{
		Addr:    ":8080",
		Handler: rootMux,
	}
})
