package main

import (
	"fmt"

	"github.com/submodule-org/submodule.go"
)

type server struct {
	config Config
	logger Logger
}
type Server interface {
	Start()
}

func (s *server) Start() {
	s.logger.Log(fmt.Sprintf("Starting server on %s:%d\n", s.config.Host, s.config.Port))
}

func setUpServer(config Config, logger Logger) Server {
	return &server{
		config,
		logger,
	}
}

var ServerMod = submodule.Make[Server](setUpServer, ConfigMod, LoggerMod)
