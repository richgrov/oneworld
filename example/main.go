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
	if finishedDestroying || blocks.Hardness(block.Type) == blocks.InstaBreak {
		player.Server.SetBlock(x, y, z, blocks.Block{blocks.Air, 0})
	}
}

func (*player) OnInteract(x, y, z, x1, y1, z1 int) {}

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
		chunk := oneworld.NewChunk(i%16, i/16)
		chunk.InitializeToAir()

		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				chunk.Set(x, 10, z, blocks.Block{
					Type: blocks.Stone,
				})
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
