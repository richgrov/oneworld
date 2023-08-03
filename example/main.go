package main

import (
	"github.com/richgrov/oneworld"
	"github.com/richgrov/oneworld/blocks"
	"github.com/richgrov/oneworld/level"
)

type player struct {
	oneworld.PlayerBase
	server *oneworld.Server
}

func (*player) OnChat(string) {}

func (player *player) OnDig(x, y, z int, finishedDestroying bool) {
	block := player.server.GetBlock(x, y, z)
	if finishedDestroying || blocks.Hardness(block.Type) == blocks.InstaBreak {
		player.server.SetBlock(x, y, z, blocks.Block{blocks.Air, 0})
	}
}

func (*player) OnInteract(x, y, z, x1, y1, z1 int) {}

func (player *player) OnUpdateChunkViewRange(outOfRange []level.ChunkPos, inRange []level.ChunkPos) {
	for _, pos := range outOfRange {
		chunk := player.server.Chunk(pos.X, pos.Z)
		if chunk != nil {
			chunk.RemoveObserver(player)
		}
	}

	chunksToLoad := make([]level.ChunkPos, 0)

	for _, pos := range inRange {
		chunk := player.server.Chunk(pos.X, pos.Z)
		if chunk != nil {
			chunk.AddObserver(player)
			continue
		}

		player.server.InitializeChunk(pos.X, pos.Z, player)
		chunksToLoad = append(chunksToLoad, pos)
	}

	if len(chunksToLoad) > 0 {
		player.server.LoadChunks(chunksToLoad)
	}
}

func createPlayer(baseEntity *oneworld.EntityBase, conn *oneworld.AcceptedConnection, server *oneworld.Server) *player {
	player := new(player)
	base := oneworld.NewBasePlayer(
		*baseEntity,
		conn,
		16,
		0,
		oneworld.Overworld,
		player,
	)
	player.PlayerBase = base
	player.server = server
	return player
}

func main() {
	listener, err := oneworld.NewListener("localhost:25565")
	if err != nil {
		panic(err)
	}
	go listener.Run()
	defer listener.Close()

	server, err := oneworld.NewServer(&level.McRegionLoader{
		WorldDir: "world",
	}, 32)
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
