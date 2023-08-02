package main

import (
	"github.com/richgrov/oneworld"
	"github.com/richgrov/oneworld/level"
)





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
}
