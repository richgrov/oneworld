package main

import (
	"bufio"
	"os"

	"github.com/richgrov/oneworld"
	"github.com/richgrov/oneworld/traits"
)

type ServerTrait struct{}

func (t *ServerTrait) OnPlayerJoin(event *oneworld.PlayerJoinEvent) {
	println(event.Player().Username(), "logged in")
}

func main() {
	server, err := oneworld.NewServer("localhost:25565", "world", 8, oneworld.Overworld)
	if err != nil {
		panic(err)
	}

	traits.Set(server, &ServerTrait{})

	bufio.NewReader(os.Stdin).ReadString('\n')
	server.Shutdown()
}
