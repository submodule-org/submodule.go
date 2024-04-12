package main

import "fmt"

func main() {
	s, e := ServerMod.SafeResolve()
	if e != nil {
		fmt.Printf("Resolve failed %+v\n", e)
	}

	s.Start()

}
