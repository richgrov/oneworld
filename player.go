package oneworld

import (
	"bufio"
	"net"

	"github.com/richgrov/oneworld/internal/level"
	"github.com/richgrov/oneworld/internal/protocol"
)

const packetBacklog = 32

type Player struct {
	id     int32
	server *Server

	reader              *bufio.Reader
	conn                net.Conn
	outboundPacketQueue chan []byte
	// When true, outboundPacketQueue is closed
	disconnected bool

	username string
}

func newPlayer(entityId int32, server *Server, reader *bufio.Reader, conn net.Conn, username string) *Player {
	player := &Player{
		id:     entityId,
		server: server,

		reader:              reader,
		conn:                conn,
		outboundPacketQueue: make(chan []byte, packetBacklog),
		disconnected:        false,

		username: username,
	}

	go player.writeLoop()

	return player
}

func (player *Player) Id() int32 {
	return player.id
}

func (player *Player) Teleport(x float64, y float64, z float64) {
	player.queuePacket(protocol.Marshal(protocol.PositionId, &protocol.Position{
		X:        x,
		Y:        y,
		Stance:   0,
		Z:        z,
		OnGround: false,
	}))

	// Send/generate chunks
	centerChunkX := int32(x / 16)
	centerchunkZ := int32(z / 16)
	viewDist := int32(player.server.viewDistance)

	chunksToLoad := make([]level.ChunkPos, 0)

	for chunkX := centerChunkX - viewDist; chunkX <= centerchunkZ+viewDist; chunkX++ {
		for chunkZ := centerchunkZ - viewDist; chunkZ <= centerchunkZ+viewDist; chunkZ++ {
			pos := level.ChunkPos{chunkX, chunkZ}

			chunk, ok := player.server.chunks[pos]
			if !ok {
				chunk = newChunk()
				chunk.viewers = append(chunk.viewers, player)
				player.server.chunks[pos] = chunk
				chunksToLoad = append(chunksToLoad, pos)
				continue
			}

			chunk.viewers = append(chunk.viewers, player)
			if chunk.isDataLoaded() {
				player.sendChunk(pos, chunk)
			}
		}
	}

	if len(chunksToLoad) > 0 {
		player.server.loadChunks(chunksToLoad)
	}
}

func (player *Player) sendChunk(pos level.ChunkPos, ch *chunk) {
	player.queuePacket(protocol.Marshal(protocol.PreChunkId, &protocol.PreChunk{
		ChunkX: pos.X,
		ChunkZ: pos.Z,
		Load:   true,
	}))

	player.queuePacket(protocol.Marshal(protocol.ChunkDataId, &protocol.ChunkData{
		StartX: pos.X * 16,
		StartY: 0,
		StartZ: pos.Z * 16,
		XSize:  16,
		YSize:  128,
		ZSize:  16,
		Data:   ch.data.CompressData(),
	}))
}

// Can safely be called even if Disconnect() was already called
func (player *Player) queuePacket(data []byte) {
	if !player.disconnected {
		player.outboundPacketQueue <- data
	}
}

func (player *Player) writeLoop() {
	for {
		data, ok := <-player.outboundPacketQueue
		if !ok {
			break
		}

		_, err := player.conn.Write(data)
		if err != nil {
			player.Disconnect()
		}
	}
}

// Can safely be called more than once
func (player *Player) Disconnect() {
	println("Disconnecting", player.username)
	player.conn.Close()

	if !player.disconnected {
		close(player.outboundPacketQueue)
	}
	player.disconnected = true
}
