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

// Decodes a struct's fields from a reader in the order they are declared
func Unmarshal(reader io.Reader, v any) error {
	ty := reflect.TypeOf(v).Elem()
	val := reflect.ValueOf(v).Elem()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		switch field.Kind() {
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
	if err := binary.Write(writer, binary.BigEndian, int16(len(str))); err != nil {
		return err
	}

	data := utf16.Encode([]rune(str))
	return binary.Write(writer, binary.BigEndian, data)
}
