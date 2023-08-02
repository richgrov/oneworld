package oneworld

import (
	"container/list"
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/richgrov/oneworld/internal/protocol"
	"github.com/richgrov/oneworld/internal/util"
	"github.com/richgrov/oneworld/level"
)

const protocolVersion = 14
const ticksPerSecond = 20
const messageQueueBacklog = 16

type Server struct {
	ticker *time.Ticker
	// Functions added to this channel will be invoked from the goroutine of the
	// main tick loop.
	messageQueue chan func()

	worldLoader  worldLoader
	viewDistance uint8

	// Below two variables have no server use. Only sent to client
	noiseSeed int64
	dimension Dimension

	spawnX int32
	spawnY int32
	spawnZ int32

	entities     map[int32]Entity
	nextEntityId int32

	// All the chunks on the server.
	//
	// If the map does not contain the key:
	// - The chunk is not loaded AND it is not actively being loaded
	//
	// If the map contains the key, but chunk.isDataLoaded() is false:
	// - The chunk is currently being loaded
	//
	// If the map contains the key AND chunk.isDataLoaded() is true:
	// - The chunk along with all its data is loaded and valid
	//
	// The map should never contain a key pointing to nil.
	chunks            map[level.ChunkPos]*chunk
	chunkLoadConsumer chan level.ChunkReadResult
}

type worldLoader interface {
	ReadWorldInfo() (level.WorldInfo, error)
	LoadChunks([]level.ChunkPos, chan level.ChunkReadResult)
}

type Config struct {
	ViewDistance uint8
	Dimension    Dimension // Only used by the client
	// If nil, chunk loading from disk will not occurr. See also [level.McRegionLoader]
	WorldLoader worldLoader
}

func NewServer(config *Config) (*Server, error) {
	worldInfo := level.WorldInfo{}
	if config.WorldLoader != nil {
		info, err := config.WorldLoader.ReadWorldInfo()
		if err != nil {
			return nil, fmt.Errorf("error loading world info: %w", err)
		}
		worldInfo = info
	}

	server := &Server{
		ticker:       time.NewTicker(time.Second / ticksPerSecond),
		messageQueue: make(chan func(), messageQueueBacklog),

		worldLoader:  config.WorldLoader,
		viewDistance: config.ViewDistance,

		noiseSeed: worldInfo.BiomeSeed,
		dimension: config.Dimension,

		spawnX: worldInfo.SpawnX,
		spawnY: worldInfo.SpawnY,
		spawnZ: worldInfo.SpawnZ,

		entities:     make(map[int32]Entity),
		nextEntityId: 0,

		chunks:            make(map[level.ChunkPos]*chunk),
		chunkLoadConsumer: make(chan level.ChunkReadResult, config.ViewDistance*config.ViewDistance),

	}

	return server, nil
}

func (server *Server) AddPlayer(conn *AcceptedConnection) {
	id := server.newEntityId()
	player := newPlayer(id, server, conn.reader, conn.conn, conn.Username)
	server.entities[id] = player

	player.queuePacket(&protocol.LoginPacket{
		ProtocolVersion: id,
		MapSeed:         server.noiseSeed,
		Dimension:       byte(server.dimension),
	})
	player.Teleport(float64(server.spawnX), float64(server.spawnY)+5.0, float64(server.spawnZ))

}

func (server *Server) Ticker() <-chan time.Time {
	return server.ticker.C
}

func (server *Server) Tick() {
	server.drainMessageQueue()
	server.tickEntities()
	server.addLoadedChunks()
}

func (server *Server) drainMessageQueue() {
	for {
		select {
		case message := <-server.messageQueue:
			message()
		default:
			return
		}
	}
}

func (server *Server) tickEntities() {
	for _, entity := range server.entities {
		entity.Tick()
	}
}

}

func (server *Server) addLoadedChunks() {
	for {
		select {
		case result := <-server.chunkLoadConsumer:
			chunk := server.chunks[result.Pos]

			if result.Error != nil {
				for player := range chunk.viewers {
					delete(player.viewableChunks, result.Pos)
				}
				delete(server.chunks, result.Pos)
				fmt.Printf("%s", result.Error)
				continue
			}

			// TODO: Until chunk generator is implemented, just set the chunk to air
			if result.Data.Blocks == nil {
				result.Data.InitializeToAir()
			}

			chunk.initialize(result.Data)
			for player := range chunk.viewers {
				player.sendChunk(result.Pos, chunk)
			}
		default:
			return
		}
	}
}

// Gets the next available entity ID
func (server *Server) newEntityId() int32 {
	id := server.nextEntityId
	if id == math.MaxInt32 {
		panic("entity IDs exhausted")
	}

	server.nextEntityId++
	return id
}

func (server *Server) loadChunks(positions []level.ChunkPos) {
	go server.worldLoader.LoadChunks(positions, server.chunkLoadConsumer)
}

func (server *Server) getChunkFromBlockPos(x int32, z int32) *chunk {
	ch, _ := server.chunks[level.ChunkPos{
		util.DivideAndFloorI32(x, 16),
		util.DivideAndFloorI32(z, 16),
	}]
	return ch
}

func (server *Server) GetBlock(x int32, y int32, z int32) *Block {
	ch := server.getChunkFromBlockPos(x, z)
	if ch == nil || !ch.isDataLoaded() {
		return nil
	}

	index := chunkCoordsToIndex(util.I32Abs(x%16), y, util.I32Abs(z%16))
	return &ch.blocks[index]
}

func (server *Server) SetBlock(x int32, y int32, z int32, block Block) bool {
	ch := server.getChunkFromBlockPos(x, z)
	if ch == nil || !ch.isDataLoaded() {
		return false
	}

	index := chunkCoordsToIndex(util.I32Abs(x%16), y, util.I32Abs(z%16))
	ch.blocks[index] = block

	for player := range ch.viewers {
		player.SendBlockChange(x, y, z, block.Type(), block.Data())
	}
	return true
}

// Stops all running server processes. The function will block until all
// processes have stopped.
func (server *Server) Shutdown() {
	server.ticker.Stop()
}
