package nbt

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
)

const (
	tagEnd = iota
	tagByte
	tagShort
	tagInt
	tagLong
	tagFloat
	tagDouble
	tagByteArray
	tagString
	tagList
	tagCompound
	tagIntArray
	tagLongArray
)

func Unmarshal(reader *bufio.Reader, v any) error {
	val := reflect.ValueOf(v).Elem()

	if tag, err := reader.ReadByte(); err != nil {
		return err
	} else if tag != tagCompound {
		return errors.New("expected NBT compound")
	}

	if _, err := readString(reader); err != nil {
		return err
	}

	return unmarshalCompound(reader, val)
}

func readString(reader io.Reader) (string, error) {
	var strLen uint16
	if err := binary.Read(reader, binary.BigEndian, &strLen); err != nil {
		return "", err
	}

	data := make([]byte, strLen)
	if _, err := io.ReadFull(reader, data); err != nil {
		return "", err
	}

	return string(data), nil
}

func unmarshalCompound(reader *bufio.Reader, v reflect.Value) error {
	if v.IsValid() && v.Kind() != reflect.Struct {
		return errors.New("expected struct type")
	}

	for {
		tag, err := reader.ReadByte()
		if err != nil {
			return nil
		}

		if tag == tagEnd {
			break
		}

		key, err := readString(reader)
		if err != nil {
			return err
		}

		var field reflect.Value
		if v.IsValid() {
			field = v.FieldByName(key)
		}

		if err := unmarshalValue(tag, reader, field); err != nil {
			return err
		}
	}

	return nil
}

func unmarshalList(reader *bufio.Reader, val reflect.Value) error {
	elementType, err := reader.ReadByte()
	if err != nil {
		return err
	}

	var listLen int32
	if err := binary.Read(reader, binary.BigEndian, &listLen); err != nil {
		return err
	}

	if listLen < 0 {
		listLen = 0
	}

	var list reflect.Value
	if val.IsValid() {
		list = reflect.MakeSlice(val.Type().Elem(), int(listLen), int(listLen))
	}

	for i := 0; i < int(listLen); i++ {
		var elem reflect.Value
		if list.IsValid() {
			elem = list.Index(i)
		}

		if err := unmarshalValue(elementType, reader, elem); err != nil {
			return err
		}
	}

	if val.IsValid() {
		val.Set(list)
	}
	return nil
}

func unmarshalPrimitiveArray[T any](reader io.Reader) (reflect.Value, error) {
	var arrLen int32
	if err := binary.Read(reader, binary.BigEndian, &arrLen); err != nil {
		return reflect.Value{}, err
	}

	data := make([]T, arrLen)
	if byteArray, ok := any(data).([]byte); ok {
		if _, err := io.ReadFull(reader, byteArray); err != nil {
			return reflect.Value{}, err
		}
	} else {
		for i := 0; i < int(arrLen); i++ {
			var entry T
			if err := binary.Read(reader, binary.BigEndian, &entry); err != nil {
				return reflect.Value{}, err
			}
			data = append(data, entry)
		}
	}

	return reflect.ValueOf(data), nil
}

func unmarshalValue(tag byte, reader *bufio.Reader, val reflect.Value) error {
	var newVal reflect.Value

	switch tag {
	case tagByte:
		b, err := reader.ReadByte()
		if err != nil {
			return err
		}
		newVal = reflect.ValueOf(b)

	case tagShort:
		var value int16
		if err := binary.Read(reader, binary.BigEndian, &value); err != nil {
			return err
		}
		newVal = reflect.ValueOf(value)

	case tagInt:
		var value int32
		if err := binary.Read(reader, binary.BigEndian, &value); err != nil {
			return err
		}
		newVal = reflect.ValueOf(value)

	case tagLong:
		var value int64
		if err := binary.Read(reader, binary.BigEndian, &value); err != nil {
			return err
		}
		newVal = reflect.ValueOf(value)

	case tagFloat:
		var value float32
		if err := binary.Read(reader, binary.BigEndian, &value); err != nil {
			return err
		}
		newVal = reflect.ValueOf(value)

	case tagDouble:
		var value float64
		if err := binary.Read(reader, binary.BigEndian, &value); err != nil {
			return err
		}
		newVal = reflect.ValueOf(value)

	case tagByteArray:
		arr, err := unmarshalPrimitiveArray[byte](reader)
		if err != nil {
			return err
		}
		newVal = arr

	case tagString:
		str, err := readString(reader)
		if err != nil {
			return err
		}
		newVal = reflect.ValueOf(str)

	case tagList:
		return unmarshalList(reader, val)

	case tagCompound:
		return unmarshalCompound(reader, val)

	case tagIntArray:
		arr, err := unmarshalPrimitiveArray[int32](reader)
		if err != nil {
			return err
		}
		newVal = arr

	case tagLongArray:
		arr, err := unmarshalPrimitiveArray[int64](reader)
		if err != nil {
			return err
		}
		newVal = arr

	default:
		return errors.New("invalid tag type")
	}

	if val.IsValid() {
		if val.Type() != newVal.Type() {
			return errors.New("type mismatch")
		}

		val.Set(newVal)
	}

	return nil
}