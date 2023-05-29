package oneworld

import (
	"bufio"
	"fmt"
	"net"
	"reflect"
	"strings"

	"github.com/richgrov/oneworld/internal/level"
	"github.com/richgrov/oneworld/internal/protocol"
	"github.com/richgrov/oneworld/traits"
)

const packetBacklog = 32

type Player struct {
	id        int32
	server    *Server
	traitData *traits.TraitData

	reader              *bufio.Reader
	conn                net.Conn
	inboundPacketQueue  chan any
	outboundPacketQueue chan []byte
	// When true, outboundPacketQueue is closed
	disconnected bool

	username string

	viewableChunks map[level.ChunkPos]*chunk
}

func newPlayer(entityId int32, server *Server, reader *bufio.Reader, conn net.Conn, username string) *Player {
	player := &Player{
		id:     entityId,
		server: server,
		traitData: traits.NewData(
			reflect.TypeOf(&ChatEvent{}),
			reflect.TypeOf(&CommandEvent{}),
		),

		reader:              reader,
		conn:                conn,
		inboundPacketQueue:  make(chan any, packetBacklog),
		outboundPacketQueue: make(chan []byte, packetBacklog),
		disconnected:        false,

		username: username,

		viewableChunks: make(map[level.ChunkPos]*chunk),
	}

	go player.readLoop()
	go player.writeLoop()

	server.Repeat(func() int {
		player.queuePacket(protocol.Marshal(protocol.KeepAliveId, &protocol.KeepAlive{}))

		if !player.disconnected {
			return 20 * 20
		} else {
			return 0
		}
	})

	return player
}

func (player *Player) Id() int32 {
	return player.id
}

func (player *Player) TraitData() *traits.TraitData {
	return player.traitData
}

func (player *Player) Tick() {
processPackets:
	for {
		select {
		case packet := <-player.inboundPacketQueue:
			player.handlePacket(packet)
		default:
			break processPackets
		}
	}
}

// Teleports the player to the speicified coordinates. Will automatically
// load/unload chunks as needed.
//
// This function is also used to spawn the player into the world after they
// login.
func (player *Player) Teleport(x float64, y float64, z float64) {
	player.queuePacket(protocol.Marshal(protocol.PositionId, &protocol.Position{
		X:        x,
		Y:        y,
		Stance:   0,
		Z:        z,
		OnGround: false,
	}))

	// Determine all the chunks that will need to be loaded at the player's new
	// position
	centerChunkX := int32(x / 16)
	centerChunkZ := int32(z / 16)
	viewDist := int32(player.server.viewDistance)

	chunksInNewView := make(map[level.ChunkPos]bool)
	for chunkX := centerChunkX - viewDist; chunkX <= centerChunkZ+viewDist; chunkX++ {
		for chunkZ := centerChunkZ - viewDist; chunkZ <= centerChunkZ+viewDist; chunkZ++ {
			chunksInNewView[level.ChunkPos{chunkX, chunkZ}] = true
		}
	}

	// Unload all the old chunks that are no longer in view
	for pos, chunk := range player.viewableChunks {
		if _, ok := chunksInNewView[pos]; !ok {
			delete(chunk.viewers, player)
			delete(player.viewableChunks, pos)
			player.queuePacket(protocol.Marshal(protocol.PreChunkId, &protocol.PreChunk{
				ChunkX: pos.X,
				ChunkZ: pos.Z,
				Load:   false,
			}))
		}
	}

	chunksToLoad := make([]level.ChunkPos, 0)

	for pos, _ := range chunksInNewView {
		chunk, ok := player.server.chunks[pos]
		if ok {
			// The chunk is already loaded or it is being loaded
			player.viewableChunks[pos] = chunk

			// If the chunk is loaded and the player was not previously viewing
			// it, send it now
			if _, ok := chunk.viewers[player]; !ok && chunk.isDataLoaded() {
				player.sendChunk(pos, chunk)
			}

			chunk.viewers[player] = true
			continue
		}

		// Initialize the chunk and queue it to load
		chunk = newChunk()
		chunk.viewers[player] = true
		player.server.chunks[pos] = chunk
		player.viewableChunks[pos] = chunk
		chunksToLoad = append(chunksToLoad, pos)
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
		XSize:  15,
		YSize:  127,
		ZSize:  15,
		Data:   ch.data.CompressData(),
	}))
}

// Can safely be called even if Disconnect() was already called
func (player *Player) queuePacket(data []byte) {
	if !player.disconnected {
		player.outboundPacketQueue <- data
	}
}

func (player *Player) handlePacket(packet any) {
	switch pkt := packet.(type) {
	case *protocol.Chat:
		if strings.HasPrefix(pkt.Message, "/") {
			traits.CallEvent(player.traitData, &CommandEvent{
				player:  player,
				Command: strings.TrimPrefix(pkt.Message, "/"),
			})
		} else {
			traits.CallEvent(player.traitData, &ChatEvent{
				player:  player,
				Message: pkt.Message,
			})
		}
	}
}

func (player *Player) readLoop() {
	defer player.Disconnect()

loop:
	for {
		id, err := player.reader.ReadByte()
		if err != nil {
			fmt.Printf("%s\n", err)
			break loop
		}

		var packet any
		switch id {
		case protocol.KeepAliveId:
			packet = &protocol.KeepAlive{}
		case protocol.ChatId:
			packet = &protocol.Chat{}
		case protocol.GroundedId:
			packet = &protocol.Grounded{}
		case protocol.PositionId:
			packet = &protocol.Position{}
		case protocol.LookId:
			packet = &protocol.Look{}
		case protocol.LookMoveId:
			packet = &protocol.LookMove{}
		default:
			fmt.Printf("unsupported packet id %d\n", id)
			break loop
		}

		if err := protocol.Unmarshal(player.reader, packet); err != nil {
			fmt.Printf("%s\n", err)
			break loop
		}
		player.inboundPacketQueue <- packet
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

func (player *Player) Username() string {
	return player.username
}

func (player *Player) Message(message string) {
	player.queuePacket(protocol.Marshal(protocol.ChatId, &protocol.Chat{
		Message: message,
	}))
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
