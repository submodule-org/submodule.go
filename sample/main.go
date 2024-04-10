package main

import "fmt"

func main() {
	s, e := ServerMod.Resolve()
	if e != nil {
		fmt.Printf("Resolve failed %+v\n", e)
	}

	s.Start()

	fn, _ := fn.Resolve()
	x := fn("ehllo")

	fmt.Printf("x: %s\n", x)
}
