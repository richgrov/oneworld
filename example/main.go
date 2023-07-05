package main

import (
	"bufio"
	"os"

	"github.com/richgrov/oneworld"
	"github.com/richgrov/oneworld/level"
	"github.com/richgrov/oneworld/traits"
)

type ServerTrait struct{}

func (t *ServerTrait) OnPlayerJoin(event *oneworld.PlayerJoinEvent) {
	println(event.Player().Username(), "logged in")
	traits.Set(event.Player(), &PlayerTrait{})
}

type PlayerTrait struct{}

func (t *PlayerTrait) OnChat(event *oneworld.ChatEvent) {
	event.Player().Message(event.Player().Username() + ": " + event.Message)
}

func (t *PlayerTrait) OnCommand(event *oneworld.CommandEvent) {
	event.Player().Message("/" + event.Command)
}

func main() {
	server, err := oneworld.NewServer(&oneworld.Config{
		Address:      "localhost:25565",
		WorldDir:     "world",
		ViewDistance: 8,
		Dimension:    oneworld.Overworld,
		WorldLoader:  &level.McRegionLoader{"world"},
	})

	if err != nil {
		panic(err)
	}

	traits.Set(server, &ServerTrait{})

	bufio.NewReader(os.Stdin).ReadString('\n')
	server.Shutdown()
}
