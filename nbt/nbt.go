package nbt

import (
	"bufio"
	"encoding/binary"
	"fmt"
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
	if tag, err := reader.ReadByte(); err != nil {
		return err
	} else if tag != tagCompound {
		return fmt.Errorf("expected root tag to be compound (%d), but got %d", tagCompound, tag)
	}

	if _, err := readString(reader); err != nil {
		return err
	}

	return unmarshalCompound(reader, reflect.ValueOf(v))
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
	if v.IsValid() {
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}

		if v.Kind() != reflect.Struct {
			return fmt.Errorf("tried to unmarshal compound into %s", v.Type().String())
		}
	}

	unmarshalledFields := 0

	for {
		tag, err := reader.ReadByte()
		if err != nil {
			return err
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
			unmarshalledFields++
		}

		if err := unmarshalValue(tag, reader, field); err != nil {
			return fmt.Errorf("in field %s: %w", key, err)
		}
	}

	if v.IsValid() && unmarshalledFields < v.NumField() {
		return fmt.Errorf("struct %s has %d fields but NBT compound only had %d", v.Type().String(), v.NumField(), unmarshalledFields)
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
		list = reflect.MakeSlice(val.Type(), int(listLen), int(listLen))

		if val.Type() != list.Type() {
			return fmt.Errorf("tried to assign %s to %s", list.Type().String(), val.Type().String())
		}
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
	if err := binary.Read(reader, binary.BigEndian, data); err != nil {
		return reflect.Value{}, err
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
		return fmt.Errorf("unsupported NBT tag %d", tag)
	}

	if val.IsValid() {
		if val.Type() != newVal.Type() {
			return fmt.Errorf("tried to assign %s to %s", newVal.Type().String(), val.Type().String())
		}

		val.Set(newVal)
	}

	return nil
}
