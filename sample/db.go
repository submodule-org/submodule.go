package sample

import (
	"log/slog"

	"github.com/submodule-org/submodule.go"
	"github.com/submodule-org/submodule.go/meta/mlogger"
)

type db struct {
	Logger *slog.Logger
}

type Db interface {
	Query()
}

func (db *db) Query() {
	db.Logger.Info("queried")
}

var DbMod = submodule.Resolve[Db](&db{}, mlogger.CreateLogger("db"))
