package protocol

import (
	"bufio"
	"errors"
)

var (
	ErrInvalidPacketId = errors.New("invalid packet id")
)

type InboundPacket[Self any] interface {
	Unmarshal(r *bufio.Reader) (Self, error)
}

type OutboundPacket interface {
	Marshal() []byte
}

func ExpectPacket[P InboundPacket[P]](r *bufio.Reader, packetId byte, p P) error {
	id, err := r.ReadByte()
	if err != nil {
		return err
	}

	if id != packetId {
		return ErrInvalidPacketId
	}

	if _, err := p.Unmarshal(r); err != nil {
		return err
	}

	return nil
}

func ReadNextPacket(r *bufio.Reader) (any, error) {
	id, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	switch id {
	case KeepAliveId:
		return new(KeepAlivePacket).Unmarshal(r)
	case LoginId:
		return new(LoginPacket).Unmarshal(r)
	case HandshakeId:
		return new(HandshakePacket).Unmarshal(r)
	case ChatId:
		return new(ChatPacket).Unmarshal(r)
	case SetOnGroundId:
		return new(SetOnGroundPacket).Unmarshal(r)
	case SetPositionId:
		return new(SetPositionPacket).Unmarshal(r)
	case SetAngleId:
		return new(SetAnglePacket).Unmarshal(r)
	case SetAngleAndPositionId:
		return new(SetAngleAndPositionPacket).Unmarshal(r)
	case DigId:
		return new(DigPacket).Unmarshal(r)
	case UseItemId:
		return new(UseItemPacket).Unmarshal(r)
	case AnimationId:
		return new(AnimationPacket).Unmarshal(r)
	case EntityActionId:
		return new(EntityActionPacket).Unmarshal(r)
	case CloseInventoryId:
		return new(CloseInventoryPacket).Unmarshal(r)
	case InventoryClickId:
		return new(InventoryClickPacket).Unmarshal(r)
	default:
		return nil, ErrInvalidPacketId
	}
}

const KeepAliveId = 0

type KeepAlivePacket struct{}

func (pkt *KeepAlivePacket) Unmarshal(*bufio.Reader) (*KeepAlivePacket, error) {
	return pkt, nil
}

func (pkt *KeepAlivePacket) Marshal() []byte {
	return marshal(KeepAliveId)
}

const LoginId = 1

type LoginPacket struct {
	ProtocolVersion int32
	Username        string
	MapSeed         int64
	Dimension       byte
}

func (pkt *LoginPacket) Unmarshal(r *bufio.Reader) (*LoginPacket, error) {
	reader := newPacketReader(r)
	pkt.ProtocolVersion = reader.readInt()
	pkt.Username = reader.readString(16)
	pkt.MapSeed = reader.readLong()
	pkt.Dimension = reader.readByte()
	return pkt, reader.err
}

func (pkt *LoginPacket) Marshal() []byte {
	return marshal(LoginId,
		pkt.ProtocolVersion,
		pkt.Username,
		pkt.MapSeed,
		pkt.Dimension,
	)
}

const HandshakeId = 2

type HandshakePacket struct {
	Username string
}

func (pkt *HandshakePacket) Unmarshal(r *bufio.Reader) (*HandshakePacket, error) {
	reader := newPacketReader(r)
	pkt.Username = reader.readString(16)
	return pkt, reader.err
}

func (pkt *HandshakePacket) Marshal() []byte {
	return marshal(HandshakeId, pkt.Username)
}

const ChatId = 3

type ChatPacket struct {
	Message string
}

func (pkt *ChatPacket) Unmarshal(r *bufio.Reader) (*ChatPacket, error) {
	reader := newPacketReader(r)
	pkt.Message = reader.readString(119)
	return pkt, reader.err
}

func (pkt *ChatPacket) Marshal() []byte {
	return marshal(ChatId, pkt.Message)
}

const SetOnGroundId = 10

type SetOnGroundPacket struct {
	OnGround bool
}

func (pkt *SetOnGroundPacket) Unmarshal(r *bufio.Reader) (*SetOnGroundPacket, error) {
	reader := newPacketReader(r)
	pkt.OnGround = reader.readBool()
	return pkt, reader.err
}

const SetPositionId = 11

type SetPositionPacket struct {
	X        float64
	Y        float64
	Stance   float64
	Z        float64
	OnGround bool
}

func (pkt *SetPositionPacket) Unmarshal(r *bufio.Reader) (*SetPositionPacket, error) {
	reader := newPacketReader(r)
	pkt.X = reader.readDouble()
	pkt.Y = reader.readDouble()
	pkt.Stance = reader.readDouble()
	pkt.Z = reader.readDouble()
	pkt.OnGround = reader.readBool()
	return pkt, reader.err
}

func (pkt *SetPositionPacket) Marshal() []byte {
	return marshal(11,
		pkt.X,
		pkt.Y,
		pkt.Stance,
		pkt.Z,
		pkt.OnGround,
	)
}

const SetAngleId = 12

type SetAnglePacket struct {
	Yaw      float32
	Pitch    float32
	OnGround bool
}

func (pkt *SetAnglePacket) Unmarshal(r *bufio.Reader) (*SetAnglePacket, error) {
	reader := newPacketReader(r)
	pkt.Yaw = reader.readFloat()
	pkt.Pitch = reader.readFloat()
	pkt.OnGround = reader.readBool()
	return pkt, reader.err
}

const SetAngleAndPositionId = 13

type SetAngleAndPositionPacket struct {
	X        float64
	Y        float64
	Stance   float64
	Z        float64
	Yaw      float32
	Pitch    float32
	OnGround bool
}

func (pkt *SetAngleAndPositionPacket) Unmarshal(r *bufio.Reader) (*SetAngleAndPositionPacket, error) {
	reader := newPacketReader(r)
	pkt.X = reader.readDouble()
	pkt.Y = reader.readDouble()
	pkt.Stance = reader.readDouble()
	pkt.Z = reader.readDouble()
	pkt.Yaw = reader.readFloat()
	pkt.Pitch = reader.readFloat()
	pkt.OnGround = reader.readBool()
	return pkt, reader.err
}

const DigId = 14

type DigPacket struct {
	Status byte
	X      int32
	Y      byte
	Z      int32
	Face   byte
}

func (pkt *DigPacket) Unmarshal(r *bufio.Reader) (*DigPacket, error) {
	reader := newPacketReader(r)
	pkt.Status = reader.readByte()
	pkt.X = reader.readInt()
	pkt.Y = reader.readByte()
	pkt.Z = reader.readInt()
	pkt.Face = reader.readByte()
	return pkt, reader.err
}

const UseItemId = 15

type UseItemPacket struct {
	X         int32
	Y         byte
	Z         int32
	Direction byte
	ItemId    int16
	StackSize byte
	Damage    int16
}

func (pkt *UseItemPacket) Unmarshal(r *bufio.Reader) (*UseItemPacket, error) {
	reader := newPacketReader(r)
	pkt.X = reader.readInt()
	pkt.Y = reader.readByte()
	pkt.Z = reader.readInt()
	pkt.Direction = reader.readByte()
	pkt.ItemId = reader.readShort()
	if pkt.ItemId >= 0 {
		pkt.StackSize = reader.readByte()
		pkt.Damage = reader.readShort()
	}
	return pkt, reader.err
}

const AnimationId = 18

type AnimationPacket struct {
	EntityId int32
	Animate  byte
}

func (pkt *AnimationPacket) Unmarshal(r *bufio.Reader) (*AnimationPacket, error) {
	reader := newPacketReader(r)
	pkt.EntityId = reader.readInt()
	pkt.Animate = reader.readByte()
	return pkt, reader.err
}

const EntityActionId = 19

type EntityAction byte

const (
	ActionStartSneak EntityAction = iota + 1
	ActionStopSneak
	ActionStopSleep
)

type EntityActionPacket struct {
	EntityId int32
	State    EntityAction
}

func (pkt *EntityActionPacket) Unmarshal(r *bufio.Reader) (*EntityActionPacket, error) {
	reader := newPacketReader(r)
	pkt.EntityId = reader.readInt()
	pkt.State = EntityAction(reader.readByte())
	return pkt, reader.err
}

const PreChunkId = 50

type PreChunkPacket struct {
	ChunkX int32
	ChunkZ int32
	// true to load, false to unload
	Load bool
}

func (pkt *PreChunkPacket) Marshal() []byte {
	return marshal(50,
		pkt.ChunkX,
		pkt.ChunkZ,
		pkt.Load,
	)
}

const ChunkDataId = 51

type ChunkDataPacket struct {
	StartX int32
	StartY int16
	StartZ int32
	XSize  byte
	YSize  byte
	ZSize  byte
	Data   []byte
}

func (pkt *ChunkDataPacket) Marshal() []byte {
	return marshal(ChunkDataId,
		pkt.StartX,
		pkt.StartY,
		pkt.StartZ,
		pkt.XSize,
		pkt.YSize,
		pkt.ZSize,
		pkt.Data,
	)
}

const BlockChangeId = 53

type BlockChangePacket struct {
	X    int32
	Y    byte
	Z    int32
	Type byte
	Data byte
}

func (pkt *BlockChangePacket) Marshal() []byte {
	return marshal(BlockChangeId,
		pkt.X,
		pkt.Y,
		pkt.Z,
		pkt.Type,
		pkt.Data,
	)
}

const CloseInventoryId = 101

type CloseInventoryPacket struct {
	WindowId byte
}

func (pkt *CloseInventoryPacket) Unmarshal(r *bufio.Reader) (*CloseInventoryPacket, error) {
	reader := newPacketReader(r)
	pkt.WindowId = reader.readByte()
	return pkt, reader.err
}

const InventoryClickId = 102

type InventoryClickPacket struct {
	WindowId   byte
	Slot       int16
	ClickType  byte
	Action     int16
	ShiftClick bool
	ItemId     int16
	StackSize  byte
	Damage     int16
}

func (pkt *InventoryClickPacket) Unmarshal(r *bufio.Reader) (*InventoryClickPacket, error) {
	reader := newPacketReader(r)
	pkt.WindowId = reader.readByte()
	pkt.Slot = reader.readShort()
	pkt.ClickType = reader.readByte()
	pkt.Action = reader.readShort()
	pkt.ShiftClick = reader.readBool()
	pkt.ItemId = reader.readShort()
	if pkt.ItemId >= 0 {
		pkt.StackSize = reader.readByte()
		pkt.Damage = reader.readShort()
	}
	return pkt, reader.err
}
