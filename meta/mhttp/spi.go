package mhttp

import "net/http"

type IntegrateWithHttpServer interface {
	AdaptToHTTPHandler(rootMux *http.ServeMux)
}
