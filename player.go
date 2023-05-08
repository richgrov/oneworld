package oneworld

import (
	"bufio"
	"net"
)

const packetBacklog = 32

type Player struct {
	id int32

	reader              *bufio.Reader
	conn                net.Conn
	outboundPacketQueue chan []byte
	// When true, outboundPacketQueue is closed
	disconnected bool

	username string
}

func newPlayer(entityId int32, reader *bufio.Reader, conn net.Conn, username string) *Player {
	player := &Player{
		id: entityId,

		reader:              reader,
		conn:                conn,
		outboundPacketQueue: make(chan []byte, packetBacklog),
		disconnected:        false,

		username: username,
	}

	go player.writeLoop()

	return player
}

func (player *Player) Id() int32 {
	return player.id
}

// Can safely be called even if Disconnect() was already called
func (player *Player) queuePacket(data []byte) {
	if !player.disconnected {
		player.outboundPacketQueue <- data
	}
}

func (player *Player) writeLoop() {
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

// Can safely be called more than once
func (player *Player) Disconnect() {
	println("Disconnecting", player.username)
	player.conn.Close()

	if !player.disconnected {
		close(player.outboundPacketQueue)
	}
	player.disconnected = true
}
