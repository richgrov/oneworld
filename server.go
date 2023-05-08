package oneworld

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/richgrov/oneworld/internal/protocol"
)

const protocolVersion = 14
const ticksPerSecond = 20

type Server struct {
	listener net.Listener
	// Send any value in this channel to terminate the tick loop
	tickLoopStopper chan byte
	// All running server processes will add to this wait group. When .Wait()
	// returns, all server processes have stopped.
	shutdownQueue sync.WaitGroup
}

// Creates a new server instance. The tick loop is started in the background,
// meaining this function will not block.
func NewServer(address string) (*Server, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener:        listener,
		tickLoopStopper: make(chan byte),
		shutdownQueue:   sync.WaitGroup{},
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

			println(username, "logged in")
		}()
	}
}

// Handles the login process for a new connection. Returns the username of the
// player
func handleConnection(reader *bufio.Reader, writer io.Writer) (string, error) {
	var handshake protocol.Handshake
	if err := readPacket(reader, protocol.HandshakeId, &handshake); err != nil {
		return "", err
	}

	// Legacy auth is no longer supported, so servers always respond with
	// offline mode handshake, which is "-" for the username.
	if err := writePacket(writer, protocol.HandshakeId, &protocol.Handshake{
		Username: "-",
	}); err != nil {
		return "", err
	}

	var login protocol.Login
	if err := readPacket(reader, protocol.LoginId, &login); err != nil {
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

// Reads a specific packet and unmarshals the data
func readPacket(reader *bufio.Reader, expectedId byte, v any) error {
	if b, err := reader.ReadByte(); err != nil {
		return err
	} else if b != expectedId {
		return errors.New("unexpected id")
	}

	return protocol.Unmarshal(reader, v)
}

// Marshals and writes a packet
func writePacket(writer io.Writer, packetId byte, v any) error {
	_, err := writer.Write(protocol.Marshal(protocol.HandshakeId, v))
	return err
}

// Runs the server's main tick loop
func (server *Server) tickLoop() {
	server.shutdownQueue.Add(1)
	defer server.shutdownQueue.Done()

	ticker := time.NewTicker(time.Second / ticksPerSecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
		case <-server.tickLoopStopper:
			return
		}
	}
}

// Stops all running server processes. The function will block until all
// processes have stopped.
func (server *Server) Shutdown() {
	server.listener.Close()
	server.tickLoopStopper <- 0
	server.shutdownQueue.Wait()
}
