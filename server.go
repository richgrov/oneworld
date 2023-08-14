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
	entityTracker []indexedEntities
	chunkDiameter int
}

type indexedEntities struct {
	observers []chunkObserver
}

type chunkObserver interface {
	initializeChunk(chunkX, chunkZ int)
	unloadChunk(chunkX, chunkZ int)
	sendChunk(chunkX, chunkZ int, chunk *Chunk)
	SendBlockChange(x, y, z int, block blocks.Block)
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

	entityIndices := make([]indexedEntities, chunkDiameter*chunkDiameter)
	for i := range entityIndices {
		entityIndices[i] = indexedEntities{
			observers: make([]chunkObserver, 0),
		}
	}

	server := &Server{
		ticker:       time.NewTicker(time.Second / ticksPerSecond),
		messageQueue: make(chan func(), messageQueueBacklog),

		entities:     make(map[int32]Entity),
		nextEntityId: 0,

		chunks:        chunks,
		entityTracker: entityIndices,
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
	index := server.indexedEntities(chunkX, chunkZ)
	index.observers = append(index.observers, observer)
	observer.initializeChunk(chunkX, chunkZ)

	chunk := server.Chunk(chunkX, chunkZ)
	if chunk != nil {
		observer.sendChunk(chunkX, chunkZ, chunk)
	}
}

func (server *Server) removeChunkObserver(chunkX, chunkZ int, observer chunkObserver) {
	index := server.indexedEntities(chunkX, chunkZ)
	for i, obs := range index.observers {
		if obs == observer {
			observer.unloadChunk(chunkX, chunkZ)
			index.observers = append(index.observers[:i], index.observers[i+1:]...)
			break
		}
	}
}

func (server *Server) Chunk(chunkX, chunkZ int) *Chunk {
	if chunkX < 0 || chunkX >= server.chunkDiameter || chunkZ < 0 || chunkZ >= server.chunkDiameter {
		return nil
	}
	return server.chunks[chunkZ*server.chunkDiameter+chunkX]
}

func (server *Server) addChunk(chunkX, chunkZ int, chunk *Chunk) {
	if chunkX < 0 || chunkX >= server.chunkDiameter || chunkZ < 0 || chunkZ >= server.chunkDiameter {
		msg := fmt.Sprint("chunk out of bounds: x=", chunkX, " z=", chunkZ, " diameter=", server.chunkDiameter)
		panic(msg)
	}
	server.chunks[chunkZ*server.chunkDiameter+chunkX] = chunk
}

func (server *Server) ChunkFromBlockPos(x, z int) *Chunk {
	return server.Chunk(x/16, z/16)
}

func (server *Server) GetBlock(x, y, z int) blocks.Block {
	ch := server.ChunkFromBlockPos(x, z)
	if ch == nil {
		return blocks.Block{}
	}

	index := chunkCoordsToIndex(x%16, y, z%16)
	return ch.Blocks[index]
}

func (server *Server) SetBlock(x, y, z int, block blocks.Block) bool {
	chunkX := x / 16
	chunkZ := z / 16

	ch := server.Chunk(chunkX, chunkZ)
	if ch == nil {
		return false
	}

	index := chunkCoordsToIndex(x%16, y, z%16)
	ch.Blocks[index] = block

	for _, player := range server.indexedEntities(chunkX, chunkZ).observers {
		player.SendBlockChange(x, y, z, block)
	}
	return true
}

func (server *Server) indexedEntities(chunkX, chunkZ int) *indexedEntities {
	if chunkX < 0 || chunkX >= server.chunkDiameter || chunkZ < 0 || chunkZ >= server.chunkDiameter {
		msg := fmt.Sprint("entity index out of bounds: x=", chunkX, " z=", chunkZ, " diameter=", server.chunkDiameter)
		panic(msg)
	}
	return &server.entityTracker[chunkZ*server.chunkDiameter+chunkX]
}

// Stops all running server processes. The function will block until all
// processes have stopped.
func (server *Server) Shutdown() {
	server.ticker.Stop()
}
