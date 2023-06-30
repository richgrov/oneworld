package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
	"unicode/utf16"
)

var (
	ErrInvalidBool   = errors.New("not a valid boolean value")
	ErrInvalidString = errors.New("invalid string length")
)

type packetReader struct {
	err error
	r   io.Reader
}

func newPacketReader(r io.Reader) *packetReader {
	return &packetReader{
		err: nil,
		r:   r,
	}
}

func (reader *packetReader) readByte() byte {
	if reader.err != nil {
		return 0
	}

	var buf [1]byte
	if _, err := reader.r.Read(buf[:]); err != nil {
		reader.err = err
		return 0
	}

	return buf[0]
}

func (reader *packetReader) readBool() bool {
	if reader.err != nil {
		return false
	}

	var buf [1]byte
	if _, err := reader.r.Read(buf[:]); err != nil {
		reader.err = err
		return false
	}

	if buf[0] == 0 {
		return false
	} else if buf[0] == 1 {
		return true
	} else {
		reader.err = ErrInvalidBool
		return false
	}
}

func (reader *packetReader) readShort() int16 {
	if reader.err != nil {
		return 0
	}

	var i int16
	if err := binary.Read(reader.r, binary.BigEndian, &i); err != nil {
		reader.err = err
		return 0
	}
	return i
}

func (reader *packetReader) readInt() int32 {
	if reader.err != nil {
		return 0
	}

	var i int32
	if err := binary.Read(reader.r, binary.BigEndian, &i); err != nil {
		reader.err = err
		return 0
	}
	return i
}

func (reader *packetReader) readLong() int64 {
	if reader.err != nil {
		return 0
	}

	var i int64
	if err := binary.Read(reader.r, binary.BigEndian, &i); err != nil {
		reader.err = err
		return 0
	}
	return i
}

func (reader *packetReader) readFloat() float32 {
	if reader.err != nil {
		return 0
	}

	var f float32
	if err := binary.Read(reader.r, binary.BigEndian, &f); err != nil {
		reader.err = err
		return 0
	}
	return f
}

func (reader *packetReader) readDouble() float64 {
	if reader.err != nil {
		return 0
	}

	var f float64
	if err := binary.Read(reader.r, binary.BigEndian, &f); err != nil {
		reader.err = err
		return 0
	}
	return f
}

// Reads a string encoded according to the Minecraft protocol
func (reader *packetReader) readString(maxLen uint16) string {
	if reader.err != nil {
		return ""
	}

	var length int16
	if err := binary.Read(reader.r, binary.BigEndian, &length); err != nil {
		reader.err = err
		return ""
	}

	if length < 0 || length > int16(maxLen) {
		reader.err = ErrInvalidString
		return ""
	}

	data := make([]uint16, length)
	if err := binary.Read(reader.r, binary.BigEndian, data); err != nil {
		reader.err = err
		return ""
	}

	return string(utf16.Decode(data))
}

// Encodes a struct's fields in the order they are declared and returns the
// bytes
func marshal(packetId byte, fields ...any) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.WriteByte(packetId)

	for _, field := range fields {
		val := reflect.ValueOf(field)

		switch val.Kind() {
		case reflect.Bool:
			if val.Bool() {
				buf.WriteByte(1)
			} else {
				buf.WriteByte(0)
			}

		case reflect.Uint8:
			buf.WriteByte(byte(val.Uint()))

		case reflect.Int16:
			binary.Write(buf, binary.BigEndian, int16(val.Int()))

		case reflect.Int32:
			binary.Write(buf, binary.BigEndian, int32(val.Int()))

		case reflect.Int64:
			binary.Write(buf, binary.BigEndian, val.Int())

		case reflect.Float64:
			binary.Write(buf, binary.BigEndian, val.Float())

		case reflect.String:
			writeString(buf, val.String())

		case reflect.Slice:
			if reflect.TypeOf(field).Elem().Kind() != reflect.Uint8 {
				panic("only []byte is supported")
			}
			slice := val.Interface().([]byte)
			binary.Write(buf, binary.BigEndian, int32(len(slice)))
			buf.Write(slice)

		default:
			panic("marshal: unsupported field type")
		}
	}

	return buf.Bytes()
}

// Writes a string encoded according to the Minecraft protocol
func writeString(writer io.Writer, str string) error {
	data := utf16.Encode([]rune(str))

	if err := binary.Write(writer, binary.BigEndian, int16(len(data))); err != nil {
		return err
	}

	return binary.Write(writer, binary.BigEndian, data)
}
