package main

import (
	"net/http"

	"github.com/submodule-org/submodule.go"
)

type emptyHandler struct {
	Logger Logger
	Db     Db
}

func (h *emptyHandler) Handle() {
	h.Db.Query()
	h.Logger.Log("Empty handler")
}

func (h *emptyHandler) AdaptToHTTPHandler(m *http.ServeMux) {
	m.HandleFunc("/empty", func(w http.ResponseWriter, r *http.Request) {
		h.Handle()
		w.Write([]byte("empty"))
	})
}

var emptyHandlerRoute = submodule.Resolve(&emptyHandler{}, LoggerMod, DbMod)
