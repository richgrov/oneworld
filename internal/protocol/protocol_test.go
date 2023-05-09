package protocol_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/richgrov/oneworld/internal/protocol"
)

type TestPacket struct {
	Bool   bool
	Byte   byte
	I32    int32
	I64    int64
	String string `maxLen:"64"`
}

func TestRead(t *testing.T) {
	// Packet should be read OK
	const packetId = 2
	packet := &TestPacket{}
	encoded := bytes.NewBuffer(protocol.Marshal(packetId, packet))

	var decoded TestPacket
	if err := protocol.ReadPacket(encoded, packetId, &decoded); err != nil {
		t.Errorf("decode failed: %s", err)
	}

	// Packet with unexpected ID should return error
	encoded = bytes.NewBuffer(protocol.Marshal(packetId, packet))
	if err := protocol.ReadPacket(encoded, packetId+1, &decoded); err == nil {
		t.Errorf("decode succeeded on invalid packet ID: %+v", &decoded)
	}
}

func TestMarshal(t *testing.T) {
	// Packet should be encoded/decoded OK
	const packetId = 2
	packet := &TestPacket{
		Bool:   true,
		Byte:   10,
		I32:    100100,
		I64:    100100100100,
		String: "hello",
	}

	encoded := bytes.NewBuffer(protocol.Marshal(packetId, packet))

	if b, err := encoded.ReadByte(); err != nil {
		t.Fatalf("error reading byte: %s", err)
	} else if b != packetId {
		t.Fatalf("first byte %d didn't match packet ID %d", b, packetId)
	}

	var decoded TestPacket
	if err := protocol.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %s", err)
	}

	if !reflect.DeepEqual(packet, &decoded) {
		t.Errorf("packets not equal:\n%+v\n%+v", packet, &decoded)
	}
}

func TestStrings(t *testing.T) {
	// Strings should be encoded/decoded OK
	testStrings := []string{
		"hello", "World", "!", "", "I\u2665Special\ufe4fSymbols",
	}

	for _, str := range testStrings {
		packet := &TestPacket{
			String: str,
		}

		encoded := bytes.NewBuffer(protocol.Marshal(0, packet))
		encoded.ReadByte()

		var decoded TestPacket
		if err := protocol.Unmarshal(encoded, &decoded); err != nil {
			t.Errorf("unmarshal error: %s", err)
		}

		if !reflect.DeepEqual(packet, &decoded) {
			t.Errorf("packets not equal:\n%+v\n%+v", packet, &decoded)
		}
	}

	// String that is too long should return error
	packet := &TestPacket{
		String: strings.Repeat("a", 100),
	}

	encoded := bytes.NewBuffer(protocol.Marshal(0, packet))
	encoded.ReadByte()

	var decoded TestPacket
	if err := protocol.Unmarshal(encoded, &decoded); err == nil {
		t.Errorf("unmarshal should return error, but got '%s'", decoded.String)
	}
}
