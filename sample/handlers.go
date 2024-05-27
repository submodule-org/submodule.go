package sample

import (
	"net/http"

	"github.com/submodule-org/submodule.go"
	"github.com/urfave/cli/v2"
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

func (h *emptyHandler) AdaptToCLI(app *cli.App) {
	app.Commands = append(app.Commands, &cli.Command{
		Name:  "empty",
		Usage: "empty handler",
		Action: func(c *cli.Context) error {
			h.Handle()
			return nil
		},
	})
}

var EmptyHandlerRoute = submodule.Resolve(&emptyHandler{}, LoggerMod, DbMod)
