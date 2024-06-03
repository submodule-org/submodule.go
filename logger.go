package submodule

import (
	"log/slog"
	"os"
	"strconv"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
)

var logger = sync.OnceValue(func() *slog.Logger {

	debug := os.Getenv("SM_DEBUG")
	zapper := zap.NewNop()
	if ok, e := strconv.ParseBool(debug); ok && e == nil {
		zapper = zap.Must(zap.NewDevelopmentConfig().Build())
	}

	l := slog.New(zapslog.NewHandler(zapper.Core(), &zapslog.HandlerOptions{
		AddSource:  true,
		LoggerName: "submodule",
	}))

	return l
})
