package oneworld

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"sync"
	"time"

	"github.com/richgrov/oneworld/internal/protocol"
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

	// Below two variables have no server use. Only sent to client
	noiseSeed int64
	dimension Dimension

	entities     map[int32]Entity
	nextEntityId int32
}

// Creates a new server instance. The tick loop is started in the background,
// meaining this function will not block.
//
// `noiseSeed` is only used by the client to calculate biome colors.
// `dimension` is also only used by the client.
func NewServer(address string, noiseSeed int64, dimension Dimension) (*Server, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener:        listener,
		tickLoopStopper: make(chan byte),
		shutdownQueue:   sync.WaitGroup{},
		messageQueue:    make(chan func(), messageQueueBacklog),

		noiseSeed: noiseSeed,
		dimension: dimension,

		entities:     make(map[int32]Entity),
		nextEntityId: 0,
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
	var handshake protocol.Handshake
	if err := protocol.ReadPacket(reader, protocol.HandshakeId, &handshake); err != nil {
		return "", err
	}

	// Legacy auth is no longer supported, so servers always respond with
	// offline mode handshake, which is "-" for the username.
	if _, err := writer.Write(protocol.Marshal(protocol.HandshakeId, &protocol.Handshake{
		Username: "-",
	})); err != nil {
		return "", err
	}

	var login protocol.Login
	if err := protocol.ReadPacket(reader, protocol.LoginId, &login); err != nil {
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
	player := newPlayer(id, reader, conn, username)
	server.entities[id] = player

	player.queuePacket(protocol.Marshal(protocol.LoginId, &protocol.Login{
		ProtocolVersion: id,
		MapSeed:         server.noiseSeed,
		Dimension:       byte(server.dimension),
	}))

	println(username, "logged in")
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
		for message := range server.messageQueue {
			message()
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

// Stops all running server processes. The function will block until all
// processes have stopped.
func (server *Server) Shutdown() {
	server.listener.Close()
	server.tickLoopStopper <- 0
	server.shutdownQueue.Wait()
}
