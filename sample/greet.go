package main

import (
	"fmt"
	"submodule"
)

type GreetService struct {
	Url string
}

func (m GreetService) Hi() {
	fmt.Printf("%s Hi \n", m.Url)
}

func (m GreetService) Bye() {
	fmt.Printf("%s Bye \n", m.Url)
}

var GetGreetService = submodule.Derive(func(p struct{ Config Config }) (GreetService, error) {
	return GreetService{Url: p.Config.Host}, nil
})
