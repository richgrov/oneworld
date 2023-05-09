package level

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"io"
)

const (
	TagEnd = iota
	TagByte
	TagShort
	TagInt
	TagLong
	TagFloat
	TagDouble
	TagByteArray
	TagString
	TagList
	TagCompound
	TagIntArray
	TagLongArray
)

func readNbt(r io.Reader) (map[string]any, error) {
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(gzipReader)

	if tag, err := reader.ReadByte(); err != nil {
		return nil, err
	} else if tag != TagCompound {
		return nil, errors.New("expected NBT compound")
	}

	if _, err := readNbtString(reader); err != nil {
		return nil, err
	}

	return readNbtCompound(reader)
}

func readNbtString(reader io.Reader) (string, error) {
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

func readNbtCompound(reader *bufio.Reader) (map[string]any, error) {
	values := make(map[string]any)

	for {
		tag, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}

		if tag == TagEnd {
			break
		}

		key, err := readNbtString(reader)
		if err != nil {
			return nil, err
		}

		value, err := readNbtDynamic(tag, reader)
		if err != nil {
			return nil, err
		}
		values[key] = value
	}

	return values, nil
}

func readNbtList(reader *bufio.Reader) ([]any, error) {
	listType, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	var listLen int32
	if err := binary.Read(reader, binary.BigEndian, &listLen); err != nil {
		return nil, err
	}

	if listLen < 0 {
		listLen = 0
	}

	list := make([]any, listLen)
	for i := 0; i < int(listLen); i++ {
		value, err := readNbtDynamic(listType, reader)
		if err != nil {
			return nil, err
		}
		list = append(list, value)
	}

	return list, nil
}

func readNbtArray[T any](reader io.Reader) ([]T, error) {
	var arrLen int32
	if err := binary.Read(reader, binary.BigEndian, &arrLen); err != nil {
		return nil, err
	}

	data := make([]T, arrLen)
	if byteArray, ok := any(data).([]byte); ok {
		if _, err := io.ReadFull(reader, byteArray); err != nil {
			return nil, err
		}
	} else {
		for i := 0; i < int(arrLen); i++ {
			var entry T
			if err := binary.Read(reader, binary.BigEndian, &entry); err != nil {
				return nil, err
			}
			data = append(data, entry)
		}
	}

	return data, nil
}

func readNbtDynamic(tag byte, reader *bufio.Reader) (any, error) {
	switch tag {
	case TagByte:
		return reader.ReadByte()

	case TagShort:
		var value int16
		err := binary.Read(reader, binary.BigEndian, &value)
		return value, err

	case TagInt:
		var value int32
		err := binary.Read(reader, binary.BigEndian, &value)
		return value, err

	case TagLong:
		var value int64
		err := binary.Read(reader, binary.BigEndian, &value)
		return value, err

	case TagFloat:
		var value float32
		err := binary.Read(reader, binary.BigEndian, &value)
		return value, err

	case TagDouble:
		var value float64
		err := binary.Read(reader, binary.BigEndian, &value)
		return value, err

	case TagByteArray:
		return readNbtArray[byte](reader)

	case TagString:
		return readNbtString(reader)

	case TagList:
		return readNbtList(reader)

	case TagCompound:
		return readNbtCompound(reader)

	case TagIntArray:
		return readNbtArray[int32](reader)

	case TagLongArray:
		return readNbtArray[int64](reader)

	default:
		return nil, errors.New("invalid tag type")
	}
}
