package oneworld

import (
	"fmt"
	"math"
	"time"

	"github.com/richgrov/oneworld/blocks"
)

const protocolVersion = 14
const ticksPerSecond = 20
const messageQueueBacklog = 16

type Server struct {
	ticker *time.Ticker
	// Functions added to this channel will be invoked from the goroutine of the
	// main tick loop.
	messageQueue chan func()

	entities     map[int32]Entity
	nextEntityId int32

	chunks        []*Chunk
	chunkDiameter int
}

func NewServer(chunkDiameter int, chunks []*Chunk) (*Server, error) {
	if chunkDiameter <= 0 {
		return nil, fmt.Errorf("invalid chunk diameter %d", chunkDiameter)
	}

	if len(chunks) != chunkDiameter*chunkDiameter {
		return nil, fmt.Errorf(
			"expected %d chunks because map diameter was %d, but got %d", chunkDiameter*chunkDiameter, chunkDiameter, len(chunks),
		)
	}

	server := &Server{
		ticker:       time.NewTicker(time.Second / ticksPerSecond),
		messageQueue: make(chan func(), messageQueueBacklog),

		entities:     make(map[int32]Entity),
		nextEntityId: 0,

		chunks:        chunks,
		chunkDiameter: chunkDiameter,
	}

	return server, nil
}

func (server *Server) ChunkDiameter() int {
	return server.chunkDiameter
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

func (server *Server) addChunkObserver(chunkX, chunkZ int, observer chunkObserver) {
	chunk := server.Chunk(chunkX, chunkZ)
	if chunk != nil {
		chunk.addObserver(observer)
		return
	}

	chunk = NewChunk(chunkX, chunkZ)
	chunk.addObserver(observer)
	server.addChunk(chunk)
}

func (server *Server) removeChunkObserver(chunkX, chunkZ int, observer chunkObserver) {
	chunk := server.Chunk(chunkX, chunkZ)
	if chunk != nil {
		chunk.removeObserver(observer)
	}
}

func (server *Server) Chunk(chunkX, chunkZ int) *Chunk {
	if chunkX < 0 || chunkX >= server.chunkDiameter || chunkZ < 0 || chunkZ >= server.chunkDiameter {
		return nil
	}
	return server.chunks[chunkZ*server.chunkDiameter+chunkX]
}

func (server *Server) addChunk(chunk *Chunk) {
	if chunk.x < 0 || chunk.x >= server.chunkDiameter || chunk.z < 0 || chunk.z >= server.chunkDiameter {
		msg := fmt.Sprint("chunk out of bounds: x=", chunk.x, " z=", chunk.z, " diameter=", server.chunkDiameter)
		panic(msg)
	}
	server.chunks[chunk.z*server.chunkDiameter+chunk.x] = chunk
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
