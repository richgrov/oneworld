package protocol

const LoginId = 1

type Login struct {
	ProtocolVersion int32
	Username        string `maxLen:"16"`
	MapSeed         int64
	Dimension       byte
}

const HandshakeId = 2

type Handshake struct {
	Username string `maxLen:"16"`
}

const PositionId = 11

type Position struct {
	X        float64
	Y        float64
	Stance   float64
	Z        float64
	OnGround bool
}
