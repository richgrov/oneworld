package nbt

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
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

func Marshal(v any, tagName string, w io.Writer) error {
	if _, err := w.Write([]byte{tagCompound}); err != nil {
		return err
	}

	if err := writeString(w, tagName); err != nil {
		return err
	}

	return marshalCompound(w, reflect.ValueOf(v))
}

func writeString(w io.Writer, val string) error {
	if len(val) > math.MaxUint16 {
		return fmt.Errorf("string length %d exceeds maximum %d", len(val), math.MaxUint16)
	}

	if err := binary.Write(w, binary.BigEndian, uint16(len(val))); err != nil {
		return err
	}

	_, err := w.Write([]byte(val))
	return err
}

func marshalList(w io.Writer, v reflect.Value) error {
	if err := writeTag(w, v.Type().Elem()); err != nil {
		return err
	}

	listLen := v.Len()
	if listLen > math.MaxInt32 {
		return fmt.Errorf("list length %d is greater than maximum encodable %d", listLen, math.MaxInt32)
	}

	if err := binary.Write(w, binary.BigEndian, int32(listLen)); err != nil {
		return err
	}

	for i := 0; i < int(listLen); i++ {
		if err := marshalValue(w, v.Index(i)); err != nil {
			return err
		}
	}

	return nil
}

func marshalCompound(w io.Writer, v reflect.Value) error {
	if v.IsValid() {
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}

		if v.Kind() != reflect.Struct {
			return fmt.Errorf("tried to marshal compound into %s", v.Type().String())
		}
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		name := v.Type().Field(i).Name

		if err := writeTag(w, field.Type()); err != nil {
			return err
		}

		if err := writeString(w, name); err != nil {
			return err
		}

		if err := marshalValue(w, field); err != nil {
			return err
		}
	}

	_, err := w.Write([]byte{tagEnd})
	return err
}

func marshalValue(w io.Writer, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Uint8:
		_, err := w.Write([]byte{byte(v.Uint())})
		return err

	case reflect.Int16:
		return binary.Write(w, binary.BigEndian, int16(v.Int()))

	case reflect.Int32:
		return binary.Write(w, binary.BigEndian, int32(v.Int()))

	case reflect.Int64:
		return binary.Write(w, binary.BigEndian, v.Int())

	case reflect.Float32:
		return binary.Write(w, binary.BigEndian, float32(v.Float()))

	case reflect.Float64:
		return binary.Write(w, binary.BigEndian, v.Float())

	case reflect.Slice:
		sliceType := v.Type().Elem()
		switch sliceType.Kind() {
		case reflect.Uint8, reflect.Int32, reflect.Int64:
			listLen := v.Len()
			if listLen > math.MaxInt32 {
				return fmt.Errorf("slice length %d exceeds maximum %d", listLen, math.MaxInt32)
			}

			if err := binary.Write(w, binary.BigEndian, int32(listLen)); err != nil {
				return err
			}

			return binary.Write(w, binary.BigEndian, v.Interface())
		default:
			return marshalList(w, v)
		}

	case reflect.String:
		return writeString(w, v.String())

	case reflect.Struct:
		return marshalCompound(w, v)

	default:
		return fmt.Errorf("cannot serialize type %s", v.Type().String())
	}
}

func writeTag(w io.Writer, ty reflect.Type) error {
	var tag byte
	switch ty.Kind() {
	case reflect.Uint8:
		tag = tagByte
	case reflect.Int16:
		tag = tagShort
	case reflect.Int32:
		tag = tagInt
	case reflect.Int64:
		tag = tagLong
	case reflect.Float32:
		tag = tagFloat
	case reflect.Float64:
		tag = tagDouble
	case reflect.Slice:
		switch ty.Elem().Kind() {
		case reflect.Uint8:
			tag = tagByteArray
		case reflect.Int32:
			tag = tagIntArray
		case reflect.Int64:
			tag = tagLongArray
		default:
			tag = tagList
		}
	case reflect.String:
		tag = tagString
	case reflect.Struct:
		tag = tagCompound
	default:
		return fmt.Errorf("cannot serialize type %s", ty.String())
	}

	_, err := w.Write([]byte{tag})
	return err
}
