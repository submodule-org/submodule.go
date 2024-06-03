package sample

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/submodule-org/submodule.go"
	"github.com/submodule-org/submodule.go/meta/mlogger"
	"github.com/submodule-org/submodule.go/meta/mredis"
	"github.com/urfave/cli/v2"
)

type emptyHandler struct {
	Logger *slog.Logger
	Db     Db
	Client *mredis.RedisClient
}

func (h *emptyHandler) Handle(ctx context.Context) {
	h.Client.Info(ctx)
	h.Db.Query()
	h.Logger.DebugContext(ctx, "empty handler")
}

func (h *emptyHandler) AdaptToHTTPHandler(m *http.ServeMux) {
	m.HandleFunc("/empty", func(w http.ResponseWriter, r *http.Request) {
		h.Handle(r.Context())
		w.Write([]byte("empty"))
	})
}

func (h *emptyHandler) AdaptToCLI(app *cli.App) {
	app.Commands = append(app.Commands, &cli.Command{
		Name:  "empty",
		Usage: "empty handler",
		Action: func(c *cli.Context) error {
			h.Handle(c.Context)
			return nil
		},
	})
}

var EmptyHandlerRoute = submodule.Resolve(&emptyHandler{}, mlogger.CreateLogger("empty"), DbMod, mredis.Client)
