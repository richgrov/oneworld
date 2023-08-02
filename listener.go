package oneworld

import (
	"bufio"
	"errors"
	"net"

	"github.com/richgrov/oneworld/internal/protocol"
)

type Listener struct {
	listener            net.Listener
	acceptedConnections chan *AcceptedConnection
}

type AcceptedConnection struct {
	Username string
	reader   *bufio.Reader
	conn     net.Conn
}

func NewListener(address string) (*Listener, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	return &Listener{
		listener:            listener,
		acceptedConnections: make(chan *AcceptedConnection, 16),
	}, nil
}

func (listener *Listener) Run() {
	for {
		conn, err := listener.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			break
		} else if err != nil {
			continue
		}

		go listener.accept(conn)
	}
}

func (listener *Listener) accept(conn net.Conn) error {
	reader := bufio.NewReader(conn)

	var handshake protocol.HandshakePacket
	if err := protocol.ExpectPacket(reader, protocol.HandshakeId, &handshake); err != nil {
		return err
	}

	// Legacy auth is no longer supported, so servers always respond with
	// offline mode handshake, which is "-" for the username.
	handshakeResponse := &protocol.HandshakePacket{Username: "-"}
	if _, err := conn.Write(handshakeResponse.Marshal()); err != nil {
		return err
	}

	var login protocol.LoginPacket
	if err := protocol.ExpectPacket(reader, protocol.LoginId, &login); err != nil {
		return err
	}

	if login.ProtocolVersion != protocolVersion {
		return errors.New("invalid protocol version")
	}

	if handshake.Username != login.Username {
		return errors.New("username mismatch")
	}

	listener.acceptedConnections <- &AcceptedConnection{
		Username: handshake.Username,
		reader:   reader,
		conn:     conn,
	}
	return nil
}

func (listener *Listener) Dequeue() *AcceptedConnection {
	select {
	case player := <-listener.acceptedConnections:
		return player
	default:
		return nil
	}
}

func (listener *Listener) Close() error {
	return listener.listener.Close()
}
