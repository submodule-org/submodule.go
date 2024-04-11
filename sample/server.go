package main

import (
	"fmt"

	"github.com/submodule-org/submodule.go"
)

type server struct {
	Config   Config
	Logger   Logger
	Handlers []Handler `group:"true"` // will resolve tag here `group:"handler"`
}

type Server interface {
	Start()
}

type Handler interface {
	Handle()
}

func (s *server) Start() {
	s.Logger.Log(fmt.Sprintf("Starting server on %s:%d\n", s.Config.Host, s.Config.Port))

}

var ServerMod = submodule.Craft[Server](
	&server{},
	ConfigMod,
	LoggerMod,
	EmptyHanlderMod,
	ZeroHandler,
)
