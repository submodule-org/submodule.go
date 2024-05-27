package sample

import (
	"fmt"

	"github.com/submodule-org/submodule.go"
)

type db struct {
}

type Db interface {
	Query()
}

func (db *db) Query() {
	fmt.Printf("queried")
}

var DbMod = submodule.Resolve[Db](&db{})
