package main

import (
	"github.com/richgrov/oneworld"
	"github.com/richgrov/oneworld/level"
	"github.com/richgrov/oneworld/traits"
)

type ServerTrait struct{}

func (t *ServerTrait) OnPlayerJoin(event *oneworld.PlayerJoinEvent) {
	println(event.Player.Username(), "logged in")
	traits.Set(event.Player, &PlayerTrait{})
}

type PlayerTrait struct{}

func (t *PlayerTrait) OnChat(event *oneworld.ChatEvent) {
	event.Player.Message(event.Player.Username() + ": " + event.Message)
}

func (t *PlayerTrait) OnCommand(event *oneworld.CommandEvent) {
	event.Player.Message("/" + event.Command)
}

func main() {
	listener, err := oneworld.NewListener("localhost:25565")
	if err != nil {
		panic(err)
	}
	go listener.Run()
	defer listener.Close()

	server, err := oneworld.NewServer(&oneworld.Config{
		ViewDistance: 8,
		Dimension:    oneworld.Overworld,
		WorldLoader:  &level.McRegionLoader{"world"},
	})
	if err != nil {
		panic(err)
	}
	defer server.Shutdown()

	for range server.Ticker() {
		if player := listener.Dequeue(); player != nil {
			server.AddPlayer(player)
		}

		server.Tick()
	}

	traits.Set(server, &ServerTrait{})
}
