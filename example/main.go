package main

import (
	"bufio"
	"os"

	"github.com/richgrov/oneworld"
)

func main() {
	server, err := oneworld.NewServer("localhost:25565", 0, oneworld.Overworld)
	if err != nil {
		panic(err)
	}

	bufio.NewReader(os.Stdin).ReadString('\n')
	server.Shutdown()
}
