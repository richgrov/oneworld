package oneworld

import (
	"bufio"
	"container/list"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/richgrov/oneworld/internal/level"
	"github.com/richgrov/oneworld/internal/protocol"
	"github.com/richgrov/oneworld/internal/util"
	"github.com/richgrov/oneworld/traits"
)

const protocolVersion = 14
const ticksPerSecond = 20
const messageQueueBacklog = 16

type Server struct {
	listener net.Listener
	// Send any value in this channel to terminate the tick loop
	tickLoopStopper chan byte
	// All running server processes will add to this wait group. When .Wait()
	// returns, all server processes have stopped.
	shutdownQueue sync.WaitGroup
	// Functions added to this channel will be invoked from the goroutine of the
	// main tick loop.
	messageQueue chan func()
	traitData    *traits.TraitData

	worldDir     string
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
	chunks map[level.ChunkPos]*chunk

	currentTick int
	schedules   list.List
}

type schedule struct {
	fn      func() int
	nextRun int
}

type Config struct {
	Address      string
	WorldDir     string
	ViewDistance uint8
	Dimension    Dimension // Only used by the client
}

func NewServer(config *Config) (*Server, error) {
	listener, err := net.Listen("tcp", config.Address)
	if err != nil {
		return nil, err
	}

	data, err := level.ReadLevelData(filepath.Join(config.WorldDir, "level.dat"))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener:        listener,
		tickLoopStopper: make(chan byte),
		shutdownQueue:   sync.WaitGroup{},
		messageQueue:    make(chan func(), messageQueueBacklog),
		traitData:       traits.NewData(reflect.TypeOf(&PlayerJoinEvent{})),

		worldDir:     config.WorldDir,
		viewDistance: config.ViewDistance,

		noiseSeed: data.RandomSeed,
		dimension: config.Dimension,

		spawnX: data.SpawnX,
		spawnY: data.SpawnY,
		spawnZ: data.SpawnZ,

		entities:     make(map[int32]Entity),
		nextEntityId: 0,

		chunks: make(map[level.ChunkPos]*chunk),

		currentTick: 0,
		schedules:   list.List{},
	}

	go server.acceptLoop(listener)
	go server.tickLoop()

	return server, nil
}

// Continuously accepts new connections
func (server *Server) acceptLoop(listener net.Listener) {
	server.shutdownQueue.Add(1)
	defer server.shutdownQueue.Done()

	for {
		conn, err := listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			break
		} else if err != nil {
			continue
		}

		go func() {
			reader := bufio.NewReader(conn)
			username, err := handleConnection(reader, conn)
			if err != nil {
				fmt.Printf("Error logging in: %s\n", err)
				return
			}

			server.messageQueue <- func() {
				server.addPlayer(reader, conn, username)
			}
		}()
	}
}

// Handles the login process for a new connection. Returns the username of the
// player
func handleConnection(reader *bufio.Reader, writer io.Writer) (string, error) {
	var handshake protocol.HandshakePacket
	if err := protocol.ExpectPacket(reader, protocol.HandshakeId, &handshake); err != nil {
		return "", err
	}

	// Legacy auth is no longer supported, so servers always respond with
	// offline mode handshake, which is "-" for the username.
	handshakeResponse := &protocol.HandshakePacket{Username: "-"}
	if _, err := writer.Write(handshakeResponse.Marshal()); err != nil {
		return "", err
	}

	var login protocol.LoginPacket
	if err := protocol.ExpectPacket(reader, protocol.LoginId, &login); err != nil {
		return "", err
	}

	if login.ProtocolVersion != protocolVersion {
		return "", errors.New("invalid protocol version")
	}

	if handshake.Username != login.Username {
		return "", errors.New("username mismatch")
	}

	return handshake.Username, nil
}

func (server *Server) addPlayer(reader *bufio.Reader, conn net.Conn, username string) {
	id := server.newEntityId()
	player := newPlayer(id, server, reader, conn, username)
	server.entities[id] = player

	player.queuePacket(&protocol.LoginPacket{
		ProtocolVersion: id,
		MapSeed:         server.noiseSeed,
		Dimension:       byte(server.dimension),
	})
	player.Teleport(float64(server.spawnX), float64(server.spawnY)+10.0, float64(server.spawnZ))

	event := &PlayerJoinEvent{
		player: player,
	}
	traits.CallEvent(server.traitData, event)
}

// Runs the server's main tick loop
func (server *Server) tickLoop() {
	server.shutdownQueue.Add(1)
	defer server.shutdownQueue.Done()

	ticker := time.NewTicker(time.Second / ticksPerSecond)
	defer ticker.Stop()

	for {
		// Wait until next tick
		select {
		case <-ticker.C:
		case <-server.tickLoopStopper:
			return
		}

		// Drain message queue
	messageQueue:
		for {
			select {
			case message := <-server.messageQueue:
				message()
			default:
				break messageQueue
			}
		}

		// Tick entities
		for _, entity := range server.entities {
			entity.Tick()
		}

		// Tick schedules
		e := server.schedules.Front()
		for {
			if e == nil {
				break
			}

			sched := e.Value.(*schedule)
			if sched.nextRun == server.currentTick {
				nextRunDelay := sched.fn()

				if nextRunDelay <= 0 {
					next := e.Next()
					server.schedules.Remove(e)
					e = next
					continue
				}

				sched.nextRun += nextRunDelay
			}

			e = e.Next()
		}

		server.currentTick++
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
	go func() {
		chunks := level.LoadChunks(filepath.Join(server.worldDir, "region"), positions)

		server.messageQueue <- func() {
			for i, data := range chunks {
				chunk := server.chunks[positions[i]]

				if data.BlockData == nil {
					// If an error ocurred, remove the chunk from all memory locations
					for player := range chunk.viewers {
						delete(player.viewableChunks, positions[i])
					}
					delete(server.chunks, positions[i])
					continue
				}

				chunk.initialize(data)
				for player := range chunk.viewers {
					player.sendChunk(positions[i], chunk)
				}
			}
		}
	}()
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

// Repeatedly calls the provided function on the server's main goroutine. The
// function is first executed on the next tick. The function's return value is
// the number of ticks to wait until calling it again. If the return value is
// less than 1, the function will not be called again.
func (server *Server) Repeat(fn func() int) {
	server.schedules.PushBack(&schedule{
		fn:      fn,
		nextRun: server.currentTick + 1,
	})
}

// Stops all running server processes. The function will block until all
// processes have stopped.
func (server *Server) Shutdown() {
	server.listener.Close()
	server.tickLoopStopper <- 0
	server.shutdownQueue.Wait()
}

func (server *Server) TraitData() *traits.TraitData {
	return server.traitData
}
