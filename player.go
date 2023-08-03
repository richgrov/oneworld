package oneworld

import (
	"bufio"
	"fmt"
	"math"
	"net"
	"time"

	"github.com/richgrov/oneworld/blocks"
	"github.com/richgrov/oneworld/internal/protocol"
	"github.com/richgrov/oneworld/internal/util"
	"github.com/richgrov/oneworld/level"
)

const packetBacklog = 32

type PlayerBase struct {
	EntityBase
	Username string

	biomeSeed int64
	dimension Dimension

	reader              *bufio.Reader
	conn                net.Conn
	inboundPacketQueue  chan any
	outboundPacketQueue chan []byte
	lastKeepAliveSent   time.Time
	// When true, outboundPacketQueue is closed
	disconnected bool
	eventHandler PlayerEventHandler

	items    [45]ItemStack
	viewDist int
}

func (player *PlayerBase) OnSpawned() {
	player.queuePacket(&protocol.LoginPacket{
		ProtocolVersion: player.id,
		MapSeed:         player.biomeSeed,
		Dimension:       byte(player.dimension),
	})

	player.queuePacket(&protocol.SetPositionPacket{
		X:        player.x,
		Y:        player.y,
		Stance:   0,
		Z:        player.z,
		OnGround: false,
	})

	chunkX := int(math.Floor(player.x / 16))
	chunkZ := int(math.Floor(player.z / 16))

	viewDiameter := player.viewDist*2 + 1
	chunksToLoad := make([]level.ChunkPos, 0, viewDiameter*viewDiameter)
	for cx := util.IMax(chunkX-player.viewDist, 0); cx <= chunkX+player.viewDist; cx++ {
		for cz := util.IMax(chunkZ-player.viewDist, 0); cz <= chunkZ+player.viewDist; cz++ {
			chunksToLoad = append(chunksToLoad, level.ChunkPos{
				X: cx,
				Z: cz,
			})
		}
	}
	player.eventHandler.OnUpdateChunkViewRange([]level.ChunkPos{}, chunksToLoad)
}

func NewBasePlayer(
	base EntityBase,
	conn *AcceptedConnection,
	viewDistance int,
	biomeSeed int64,
	dimension Dimension,
	eventHandler PlayerEventHandler,
) PlayerBase {
	if viewDistance <= 0 {
		panic("view distance must be positive")
	}

	player := PlayerBase{
		EntityBase: base,
		Username:   conn.Username,

		biomeSeed: biomeSeed,
		dimension: dimension,

		reader:              conn.reader,
		conn:                conn.conn,
		inboundPacketQueue:  make(chan any, packetBacklog),
		outboundPacketQueue: make(chan []byte, packetBacklog),
		disconnected:        false,
		eventHandler:        eventHandler,

		viewDist: viewDistance,
	}

	go player.readLoop()
	go player.writeLoop()

	return player
}

func (player *PlayerBase) Tick() {
	now := time.Now()
	if now.Sub(player.lastKeepAliveSent).Seconds() > 20 {
		player.queuePacket(&protocol.KeepAlivePacket{})
		player.lastKeepAliveSent = now
	}

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
func (player *PlayerBase) Teleport(x float64, y float64, z float64) {
	player.queuePacket(&protocol.SetPositionPacket{
		X:        x,
		Y:        y,
		Stance:   0,
		Z:        z,
		OnGround: false,
	})

	chunkX := int(math.Floor(player.x / 16))
	chunkZ := int(math.Floor(player.z / 16))

	newChunkX := int(math.Floor(x / 16))
	newChunkZ := int(math.Floor(z / 16))

	chunksToUnload := make([]level.ChunkPos, 0, player.viewDist*3)
	for cx := util.IMax(chunkX-player.viewDist, 0); cx <= chunkX+player.viewDist; cx++ {
		for cz := util.IMax(chunkZ-player.viewDist, 0); cz <= chunkZ+player.viewDist; cz++ {
			canSeeChunk := util.IAbs(cx-newChunkX) <= player.viewDist && util.IAbs(cz-newChunkZ) <= player.viewDist
			if !canSeeChunk {
				chunksToUnload = append(chunksToUnload, level.ChunkPos{
					X: cx,
					Z: cz,
				})
			}
		}
	}

	chunksToLoad := make([]level.ChunkPos, 0, player.viewDist*3)
	for cx := newChunkX - player.viewDist; cx <= newChunkX+player.viewDist; cx++ {
		for cz := newChunkZ - player.viewDist; cz <= newChunkZ+player.viewDist; cz++ {
			sawChunkBefore := util.IAbs(cx-chunkX) <= player.viewDist && util.IAbs(cz-chunkZ) <= player.viewDist
			if !sawChunkBefore {
				chunksToLoad = append(chunksToLoad, level.ChunkPos{
					X: cx,
					Z: cz,
				})
			}
		}
	}

	if len(chunksToUnload) > 0 || len(chunksToLoad) > 0 {
		player.eventHandler.OnUpdateChunkViewRange(chunksToUnload, chunksToLoad)
	}
}

func (player *PlayerBase) initializeChunk(chunkX int, chunkZ int) {
	player.queuePacket(&protocol.PreChunkPacket{
		ChunkX: int32(chunkX),
		ChunkZ: int32(chunkZ),
		Load:   true,
	})
}

func (player *PlayerBase) unloadChunk(chunkX int, chunkZ int) {
	player.queuePacket(&protocol.PreChunkPacket{
		ChunkX: int32(chunkX),
		ChunkZ: int32(chunkZ),
		Load:   false,
	})
}

func (player *PlayerBase) sendChunk(chunkX int, chunkZ int, ch *Chunk) {
	player.queuePacket(&protocol.ChunkDataPacket{
		StartX: int32(chunkX * 16),
		StartY: 0,
		StartZ: int32(chunkZ * 16),
		XSize:  15,
		YSize:  127,
		ZSize:  15,
		Data:   ch.serializeToNetwork(),
	})
}

// Can safely be called even if Disconnect() was already called
func (player *PlayerBase) queuePacket(packet protocol.OutboundPacket) {
	if !player.disconnected {
		player.outboundPacketQueue <- packet.Marshal()
	}
}

func (player *PlayerBase) handlePacket(packet any) {
	switch pkt := packet.(type) {
	case *protocol.ChatPacket:
		player.eventHandler.OnChat(pkt.Message)

	case *protocol.DigPacket:
		player.eventHandler.OnDig(int(pkt.X), int(pkt.Y), int(pkt.Z), pkt.Status == 2)

	case *protocol.UseItemPacket:
		if pkt.ItemId != -1 {
			x := pkt.X
			y := int32(pkt.Y)
			z := pkt.Z

			switch pkt.Direction {
			case protocol.Under:
				y--
			case protocol.Above:
				y++
			case protocol.NegativeZ:
				z--
			case protocol.PositiveZ:
				z++
			case protocol.NegativeX:
				x--
			case protocol.PositiveX:
				x++
			}

			player.eventHandler.OnInteract(int(pkt.X), int(pkt.Y), int(pkt.Z), int(x), int(y), int(z))
		}
	}
}

func (player *PlayerBase) readLoop() {
	defer player.Disconnect()

	for {
		packet, err := protocol.ReadNextPacket(player.reader)
		if err != nil {
			fmt.Printf("%s\n", err)
			break
		}
		player.inboundPacketQueue <- packet
	}
}

func (player *PlayerBase) writeLoop() {
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

func (player *PlayerBase) SendBlockChange(x int, y int, z int, ty blocks.BlockType, data byte) {
	player.queuePacket(&protocol.BlockChangePacket{
		X:    int32(x),
		Y:    byte(y),
		Z:    int32(z),
		Type: byte(ty),
		Data: data,
	})
}

func (player *PlayerBase) Message(message string) {
	player.queuePacket(&protocol.ChatPacket{
		Message: message,
	})
}

func (player *PlayerBase) GetItemInSlot(slot byte) *ItemStack {
	return &player.items[slot]
}

func (player *PlayerBase) SetItem(slot byte, item *ItemStack) {
	player.items[slot] = *item
	player.queuePacket(&protocol.SetSlotPacket{
		WindowId:  0,
		Slot:      int16(slot),
		ItemId:    int16(item.Id),
		StackSize: item.Count,
		Damage:    int16(item.Damage),
	})
}

// Can safely be called more than once
func (player *PlayerBase) Disconnect() {
	println("Disconnecting", player.Username)
	player.conn.Close()

	if !player.disconnected {
		close(player.outboundPacketQueue)
	}
	player.disconnected = true
}

type PlayerEventHandler interface {
	OnChat(message string)
	OnInteract(clickedX, clickedY, clickedZ, newX, newY, newZ int)
	OnDig(x, y, z int, finishedDestroying bool)
	OnUpdateChunkViewRange(unload []level.ChunkPos, load []level.ChunkPos)
}
