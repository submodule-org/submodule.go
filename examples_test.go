package submodule_test

import (
	"fmt"
	"net/http"

	"github.com/submodule-org/submodule.go"
)

type Config struct {
	Host     string
	Port     int
	LogLevel string
}

func LoadConfig() Config {
	// load config from ENV etc
	return Config{
		Host:     "",      // value from env or default value
		Port:     0,       // value from env or default value
		LogLevel: "debug", // value from env or default value
	}
}

// ConfigMod will be the singleton container for config value
var ConfigMod = submodule.Make[Config](LoadConfig)

type logger struct {
	LogLevel string
}

type Logger interface {
	Log(msg ...string)
}

func (l *logger) Log(msgs ...string) {
	// log implementation with log level
}

var LoggerMod = submodule.Make[Logger](func(config Config) Logger {
	return &logger{
		LogLevel: config.LogLevel,
	}
}, ConfigMod)

type server struct {
	Config Config
	Logger Logger
}

func (s *server) Start() {
	go func() {
		http.ListenAndServe(fmt.Sprintf("%s:%d", s.Config.Host, s.Config.Port), nil)
	}()
}

var ServerMod = submodule.Resolve(&server{}, ConfigMod, LoggerMod)

func main() {
	server := ServerMod.Resolve()
	server.Start()
}
