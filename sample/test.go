package main

import (
	"fmt"

	"github.com/submodule-org/submodule.go"
)

type byeStruct struct {
	Prefix string
}

type ByeService interface {
	Hi() string
	Bye()
}

func (m byeStruct) Hi() string {
	fmt.Printf("%s > Hi \n", m.Prefix)
	return m.Prefix
}

func (m byeStruct) Bye() {
	fmt.Printf("%s > Bye \n", m.Prefix)
}

var _ = submodule.Derive(func(p struct {
	Config
	Env
	HiService
}) (ByeService, error) {
	p.HiService.Bye()
	return byeStruct{Prefix: p.Config.Host}, nil
})
