package protocol

import (
	"bytes"
	"reflect"
	"testing"
	"unicode/utf16"
)

type TestPacket struct {
	Field1 byte
	Field2 int32
	Field3 int64
	Field4 string `maxLen:"5"`
}

func TestRead(t *testing.T) {
	const packetId = 2
	packet := &TestPacket{}
	encoded := bytes.NewBuffer(Marshal(packetId, packet))

	var decoded TestPacket
	if err := ReadPacket(encoded, packetId, &decoded); err != nil {
		t.Errorf("decode failed: %s", err)
	}

	encoded = bytes.NewBuffer(Marshal(packetId, packet))
	if err := ReadPacket(encoded, packetId+1, &decoded); err == nil {
		t.Errorf("decode succeeded on invalid packet ID: %+v", &decoded)
	}
}

func TestMarshal(t *testing.T) {
	const packetId = 2
	packet := &TestPacket{
		Field1: 10,
		Field2: 100100,
		Field3: 100100100100,
		Field4: "hello",
	}

	encoded := bytes.NewBuffer(Marshal(packetId, packet))

	if b, err := encoded.ReadByte(); err != nil {
		t.Fatalf("error reading byte: %s", err)
	} else if b != packetId {
		t.Fatalf("first byte %d didn't match packet ID %d", b, packetId)
	}

	var decoded TestPacket
	if err := Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %s", err)
	}

	if !reflect.DeepEqual(packet, &decoded) {
		t.Errorf("packets not equal:\n%+v\n%+v", packet, &decoded)
	}
}

func TestStringMax(t *testing.T) {
	packet := &TestPacket{
		Field4: "aaaaaa",
	}

	encoded := bytes.NewBuffer(Marshal(0, packet))
	encoded.ReadByte()

	var decoded TestPacket
	if err := Unmarshal(encoded, &decoded); err == nil {
		t.Errorf("unmarshal should return error, but got '%s'", decoded.Field4)
	}
}

func TestStrings(t *testing.T) {
	strings := []string{
		"hello", "World", "!", "", "I\u2665Special\ufe4fSymbols",
	}

	for _, str := range strings {
		buf := bytes.NewBuffer(make([]byte, 0))
		if err := writeString(buf, str); err != nil {
			t.Fatalf("writing '%s' to local buffer returned error: %s", str, err)
		}

		numCodeUnits := len(utf16.Encode([]rune(str)))

		decoded, err := readString(buf, uint16(numCodeUnits))
		if err != nil {
			t.Fatalf("reading encoded string '%s' failed: %s", str, err)
		}

		if str != decoded {
			t.Errorf("decoded string '%s' doesn't match encoded '%s'", decoded, str)
		}
	}

	buf := bytes.NewBuffer(make([]byte, 0))
	writeString(buf, "four")
	str, err := readString(buf, 2)
	if err == nil {
		t.Errorf("decoding too-long string should have returned error, but got '%s'", str)
	}
}
