package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
	"strconv"
	"unicode/utf16"
)

// Reads a specific packet and unmarshals the data
func ReadPacket(reader io.Reader, expectedId byte, v any) error {
	var b [1]byte
	if _, err := reader.Read(b[:]); err != nil {
		return err
	} else if b[0] != expectedId {
		return errors.New("unexpected id")
	}

	return Unmarshal(reader, v)
}

// Decodes a struct's fields from a reader in the order they are declared
func Unmarshal(reader io.Reader, v any) error {
	ty := reflect.TypeOf(v).Elem()
	val := reflect.ValueOf(v).Elem()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		switch field.Kind() {
		case reflect.Bool:
			var buf [1]byte
			if _, err := reader.Read(buf[:]); err != nil {
				return err
			}
			// Minecraft protocol considers any value that is not zero as true
			field.SetBool(buf[0] != 0)

		case reflect.Uint8:
			var buf [1]byte
			if _, err := reader.Read(buf[:]); err != nil {
				return err
			}
			field.SetUint(uint64(buf[0]))

		case reflect.Int32:
			var i int32
			if err := binary.Read(reader, binary.BigEndian, &i); err != nil {
				return err
			}
			field.SetInt(int64(i))

		case reflect.Int64:
			var i int64
			if err := binary.Read(reader, binary.BigEndian, &i); err != nil {
				return err
			}
			field.SetInt(i)

		case reflect.Float64:
			var f float64
			if err := binary.Read(reader, binary.BigEndian, &f); err != nil {
				return err
			}
			field.SetFloat(f)

		case reflect.String:
			tag := ty.Field(i).Tag.Get("maxLen")
			maxLen, err := strconv.ParseUint(tag, 10, 16)
			if err != nil {
				return err
			}

			str, err := readString(reader, uint16(maxLen))
			if err != nil {
				return err
			}

			field.SetString(str)
		}
	}

	return nil
}

// Encodes a struct's fields in the order they are declared and returns the
// bytes
func Marshal(packetId byte, v any) []byte {
	val := reflect.ValueOf(v).Elem()
	buf := bytes.NewBuffer(make([]byte, 0))
	binary.Write(buf, binary.BigEndian, packetId)

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		switch field.Kind() {
		case reflect.Bool:
			if field.Bool() {
				buf.WriteByte(1)
			} else {
				buf.WriteByte(0)
			}

		case reflect.Uint8:
			buf.WriteByte(byte(field.Uint()))

		case reflect.Int32:
			binary.Write(buf, binary.BigEndian, int32(field.Int()))

		case reflect.Int64:
			binary.Write(buf, binary.BigEndian, field.Int())

		case reflect.Float64:
			binary.Write(buf, binary.BigEndian, field.Float())

		case reflect.String:
			writeString(buf, field.String())
		}
	}

	return buf.Bytes()
}

// Reads a string encoded according to the Minecraft protocol
func readString(reader io.Reader, maxLen uint16) (string, error) {
	var length int16
	if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
		return "", err
	}

	if length < 0 || length > int16(maxLen) {
		return "", errors.New("string length is invalid")
	}

	data := make([]uint16, length)
	if err := binary.Read(reader, binary.BigEndian, data); err != nil {
		return "", err
	}

	return string(utf16.Decode(data)), nil
}

// Writes a string encoded according to the Minecraft protocol
func writeString(writer io.Writer, str string) error {
	data := utf16.Encode([]rune(str))

	if err := binary.Write(writer, binary.BigEndian, int16(len(data))); err != nil {
		return err
	}

	return binary.Write(writer, binary.BigEndian, data)
}
