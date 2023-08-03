package oneworld

import (
	"fmt"
	"math"
	"time"

	"github.com/richgrov/oneworld/blocks"
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

	worldLoader worldLoader

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
	chunks            []*Chunk
	chunkDiameter     int
	chunkLoadConsumer chan level.ChunkReadResult
}

type worldLoader interface {
	ReadWorldInfo() (level.WorldInfo, error)
	LoadChunks([]level.ChunkPos, chan level.ChunkReadResult)
}

func NewServer(worldLoader worldLoader, chunkDiameter int) (*Server, error) {
	if chunkDiameter <= 0 {
		panic("chunk diameter must be positive")
	}
	server := &Server{
		ticker:       time.NewTicker(time.Second / ticksPerSecond),
		messageQueue: make(chan func(), messageQueueBacklog),

		worldLoader: worldLoader,

		entities:     make(map[int32]Entity),
		nextEntityId: 0,

		chunks:            make([]*Chunk, chunkDiameter*chunkDiameter),
		chunkDiameter:     chunkDiameter,
		chunkLoadConsumer: make(chan level.ChunkReadResult, 32),
	}

	return server, nil
}

func (server *Server) AddEntity(entity Entity) {
	server.entities[entity.Id()] = entity
	entity.OnSpawned()
}

func (server *Server) AllocateEntity(x, y, z float64) EntityBase {
	id := server.nextEntityId
	if id == math.MaxInt32 {
		panic("entity IDs exhausted")
	}
	server.nextEntityId++

	return EntityBase{
		id: id,
		x:  x,
		y:  y,
		z:  z,
	}
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

func (server *Server) addLoadedChunks() {
	for {
		select {
		case result := <-server.chunkLoadConsumer:
			chunk := server.Chunk(result.ChunkX, result.ChunkZ)

			if result.Error != nil {
				fmt.Printf("%s", result.Error)
				server.SetChunk(result.ChunkX, result.ChunkZ, nil)
				continue
			}

			// TODO: Until chunk generator is implemented, just set the chunk to air
			if result.Data.Blocks == nil {
				result.Data.InitializeToAir()
			}
			chunk.blocks = result.Data.Blocks
			chunk.blockLight = result.Data.BlockLight
			chunk.skyLight = result.Data.SkyLight

			for _, player := range chunk.observers {
				player.sendChunk(result.ChunkX, result.ChunkZ, chunk)
			}
		default:
			return
		}
	}
}

func (server *Server) InitializeChunk(chunkX, chunkZ int, observers ...chunkObserver) {
	chunk := &Chunk{
		x:         chunkX,
		z:         chunkZ,
		observers: make([]chunkObserver, 0, len(observers)),
	}
	for _, observer := range observers {
		chunk.AddObserver(observer)
	}
	server.SetChunk(chunkX, chunkZ, chunk)
}

func (server *Server) LoadChunks(positions []level.ChunkPos) {
	go server.worldLoader.LoadChunks(positions, server.chunkLoadConsumer)
}

func (server *Server) Chunk(chunkX, chunkZ int) *Chunk {
	if chunkX < 0 || chunkX >= server.chunkDiameter || chunkZ < 0 || chunkZ >= server.chunkDiameter {
		return nil
	}
	return server.chunks[chunkZ*server.chunkDiameter+chunkX]
}

func (server *Server) SetChunk(chunkX, chunkZ int, val *Chunk) {
	if chunkX < 0 || chunkX >= server.chunkDiameter || chunkZ < 0 || chunkZ >= server.chunkDiameter {
		msg := fmt.Sprint("chunk out of bounds: x =", chunkX, "z =", chunkZ, "diameter =", server.chunkDiameter)
		panic(msg)
	}
	server.chunks[chunkZ*server.chunkDiameter+chunkX] = val
}

func (server *Server) ChunkFromBlockPos(x, z int) *Chunk {
	return server.Chunk(x/16, z/16)
}

func (server *Server) GetBlock(x, y, z int) blocks.Block {
	ch := server.ChunkFromBlockPos(x, z)
	if ch == nil || !ch.isDataLoaded() {
		return blocks.Block{}
	}

	index := chunkCoordsToIndex(x%16, y, z%16)
	return ch.blocks[index]
}

func (server *Server) SetBlock(x, y, z int, block blocks.Block) bool {
	ch := server.ChunkFromBlockPos(x, z)
	if ch == nil || !ch.isDataLoaded() {
		return false
	}

	index := chunkCoordsToIndex(x%16, y, z%16)
	ch.blocks[index] = block

	for _, player := range ch.observers {
		player.SendBlockChange(x, y, z, block.Type, block.Data)
	}
	return true
}

// Stops all running server processes. The function will block until all
// processes have stopped.
func (server *Server) Shutdown() {
	server.ticker.Stop()
}
