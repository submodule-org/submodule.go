package main

import "github.com/submodule-org/submodule.go"

type emptyHandler struct {
	Logger Logger
	Db     Db
}

func (h *emptyHandler) Handle() {
	h.Db.Query()
	h.Logger.Log("Empty handler")

}

var EmptyHanlderMod = submodule.Craft[Handler](
	&emptyHandler{},
	LoggerMod,
	DbMod,
)

var ZeroHandler = submodule.Craft[Handler](
	&emptyHandler{},
	LoggerMod,
	DbMod,
)
