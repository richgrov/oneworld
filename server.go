package oneworld

import (
	"fmt"
	"math"
	"time"

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
	chunks            map[level.ChunkPos]*Chunk
	chunkLoadConsumer chan level.ChunkReadResult
}

type worldLoader interface {
	ReadWorldInfo() (level.WorldInfo, error)
	LoadChunks([]level.ChunkPos, chan level.ChunkReadResult)
}

func NewServer(worldLoader worldLoader) (*Server, error) {
	server := &Server{
		ticker:       time.NewTicker(time.Second / ticksPerSecond),
		messageQueue: make(chan func(), messageQueueBacklog),

		worldLoader: worldLoader,

		entities:     make(map[int32]Entity),
		nextEntityId: 0,

		chunks:            make(map[level.ChunkPos]*Chunk),
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
			chunk := server.chunks[result.Pos]

			if result.Error != nil {
				delete(server.chunks, result.Pos)
				fmt.Printf("%s", result.Error)
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
				player.sendChunk(result.Pos.X, result.Pos.Z, chunk)
			}
		default:
			return
		}
	}
}

func (server *Server) InitializeChunk(pos level.ChunkPos, observers ...chunkObserver) {
	chunk := &Chunk{
		x:         pos.X,
		z:         pos.Z,
		observers: make([]chunkObserver, 0, len(observers)),
	}
	for _, observer := range observers {
		chunk.AddObserver(observer)
	}
	server.chunks[pos] = chunk
}

func (server *Server) LoadChunks(positions []level.ChunkPos) {
	go server.worldLoader.LoadChunks(positions, server.chunkLoadConsumer)
}

func (server *Server) GetChunk(pos level.ChunkPos) *Chunk {
	return server.chunks[pos]
}

func (server *Server) getChunkFromBlockPos(x int32, z int32) *Chunk {
	ch, _ := server.chunks[level.ChunkPos{
		util.DivideAndFloorI32(x, 16),
		util.DivideAndFloorI32(z, 16),
	}]
	return ch
}

func (server *Server) GetBlock(x int32, y int32, z int32) *blocks.Block {
	ch := server.getChunkFromBlockPos(x, z)
	if ch == nil || !ch.isDataLoaded() {
		return nil
	}

	index := chunkCoordsToIndex(util.I32Abs(x%16), y, util.I32Abs(z%16))
	return &ch.blocks[index]
}

func (server *Server) SetBlock(x int32, y int32, z int32, block blocks.Block) bool {
	ch := server.getChunkFromBlockPos(x, z)
	if ch == nil || !ch.isDataLoaded() {
		return false
	}

	index := chunkCoordsToIndex(util.I32Abs(x%16), y, util.I32Abs(z%16))
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
