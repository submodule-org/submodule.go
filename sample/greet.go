package main

import (
	"fmt"

	"github.com/submodule-org/submodule.go"
)

type hiStruct struct {
	Url string
}

type HiService interface {
	Hi() string
	Bye()
}

func (m hiStruct) Hi() string {
	fmt.Printf("%s > Hi \n", m.Url)
	return m.Url
}

func (m hiStruct) Bye() {
	fmt.Printf("%s > Bye \n", m.Url)
}

var hiProvider = submodule.Derive(func(p struct {
	Config
	Env
}) (HiService, error) {
	return hiStruct{Url: p.Config.Host}, nil
})
