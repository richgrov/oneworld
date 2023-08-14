package main

import (
	"github.com/richgrov/oneworld"
	"github.com/richgrov/oneworld/blocks"
)

type player struct {
	oneworld.PlayerBase[*oneworld.Server]
}

func (*player) OnChat(string) {}

func (player *player) OnDig(x, y, z int, finishedDestroying bool) {
	block := player.Server.GetBlock(x, y, z)
	if finishedDestroying || block.Hardness() == blocks.InstaBreak {
		player.Server.SetBlock(x, y, z, blocks.Block{})
	}
}

func (*player) OnInteractBlock(x, y, z, x1, y1, z1 int) {}
func (*player) OnInteractAir()                          {}

func createPlayer(baseEntity *oneworld.EntityBase, conn *oneworld.AcceptedConnection, server *oneworld.Server) *player {
	player := new(player)
	base := oneworld.NewBasePlayer(
		*baseEntity,
		server,
		conn,
		16,
		0,
		oneworld.Overworld,
		player,
	)
	player.PlayerBase = base
	return player
}

func main() {
	listener, err := oneworld.NewListener("localhost:25565")
	if err != nil {
		panic(err)
	}
	go listener.Run()
	defer listener.Close()

	chunks := make([]*oneworld.Chunk, 16*16)
	for i := 0; i < len(chunks); i++ {
		chunk := new(oneworld.Chunk)

		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				chunk.Set(x, 10, z, blocks.Block{
					Type: blocks.Stone,
				})

				for y := 0; y < 128; y++ {
					chunk.SetSkyLight(x, y, z, 15)
				}
			}
		}

		chunks[i] = chunk
	}

	server, err := oneworld.NewServer(16, chunks)
	if err != nil {
		panic(err)
	}
	defer server.Shutdown()

	for range server.Ticker() {
		if conn := listener.Dequeue(); conn != nil {
			base := server.AllocateEntity(0, 80, 0)
			server.AddEntity(createPlayer(&base, conn, server))
		}

		server.Tick()
	}
}
