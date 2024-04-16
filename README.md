# Simplify Service Lifecycle Management in Go with Submodule

**Effortlessly manage the lifecycle of your services in Go with Submodule**, a lightweight and versatile library designed to streamline service management. Say goodbye to complex dependency handling, configuration management, and testing challenges.

## Clear Structure, Easy Management

* **Organize services with ease:** Wrap your service creation functions and chain them together for a clear and organized structure.
* **Simplify complex dependencies:** Manage and understand intricate dependencies between services more effectively.
* **Seamlessly integrate with your framework:** Bring Submodule into your existing frameworks and utilize its power wherever you have an async function.
* **Serverless:** Initialize what function needs, not what framework wants

## Streamlined Testing

* **Flexible testing environment:** Easily change dependencies for testing purposes, promoting testability and isolation.
* **Testable code chunks:** Organize your code into smaller, testable units, facilitating robust testing.
* **Controlled lifecycle management:** Implement unit tests and integration tests with ease by controlling the lifecycle of services.

## Lightweight and Simple

* **Quick to understand and adopt:** Experience the simplicity and elegance of Submodule, even for developers new to the library.

Discover a painless way to manage the lifecycle of your services in Go with Submodule. Enhance your development workflow, improve code maintainability, and simplify testing processes.

## 💡 Usage

You can import `submodule` using:

```go
import (
    "github.com/submodule-org/submodule.go"
)
```

Then create a submodule like this:

```go
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
var ConfigMod = submodule.Provide(LoadConfig)

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

var ServerMod = submodule.Craft(&server{}, ConfigMod, LoggerMod)

func main() {
	server := ServerMod.Resolve()
	server.Start()
}
```

## 📚 Documentation
see [godoc](https://pkg.go.dev/github.com/submodule-org/submodule.go)
more examples in [submodule_test.go](module_test.go)
