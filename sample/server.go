package main

import (
	"fmt"

	"github.com/submodule-org/submodule.go"
)

type server struct {
	Config Config
	Logger Logger
}

type Server interface {
	Start()
}

func (s *server) Start() {
	s.Logger.Log(fmt.Sprintf("Starting server on %s:%d\n", s.Config.Host, s.Config.Port))
}

var ServerMod = submodule.Craft[Server](&server{}, ConfigMod, LoggerMod)
