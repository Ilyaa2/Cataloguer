package main

import (
	"Cataloguer/cmd/server"
	"fmt"
)

func main() {
	serv := server.New()
	serv.Start()
	fmt.Print("server started")
}
