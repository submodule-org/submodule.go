package main

func main() {
	gs, e := GetGreetService()
	if e != nil {
		panic(e)
	}
	gs.Hi()

}
