package nbt_test

import (
	"bufio"
	"bytes"
	"math"
	"reflect"
	"testing"

	"github.com/richgrov/oneworld/nbt"
)

// Test case inspired by htttps://wiki.vg/NBT 'bigtest.nbt'
type TestStruct struct {
	Nested     TestNested
	IntTest    int32
	ShortTest  int16
	ByteTest   byte
	StringTest string
	LongList   []int64
	DoubleTest float64
	FloatTest  float32
	LongTest   int64
	ListTest   []TestNested2
	BytesTest  []byte
	IntsTest   []int32
}

type TestNested struct {
	Egg TestNested2
	Ham TestNested2
}

type TestNested2 struct {
	Name  string
	Value float32
}

func TestMarshalUnmarshal(t *testing.T) {
	val := TestStruct{
		Nested: TestNested{
			Egg: TestNested2{
				Name:  "Eggbert",
				Value: 0.5,
			},
			Ham: TestNested2{
				Name:  "Hampus",
				Value: 0.75,
			},
		},
		IntTest:    2147483647,
		ShortTest:  math.MaxInt16,
		ByteTest:   127,
		StringTest: "Hello, world! \xC5",
		LongList:   []int64{11, 12, math.MaxInt64, math.MinInt64},
		DoubleTest: 0.49312871321823148,
		FloatTest:  0.49312871321823148,
		LongTest:   math.MaxInt64,
		ListTest: []TestNested2{
			{
				Name:  "struct #1",
				Value: 1,
			},
			{
				Name:  "struct #2",
				Value: 2,
			},
		},
		BytesTest: []byte{0xFF, 0xA7},
		IntsTest:  []int32{123, math.MaxInt32, math.MinInt32, 321},
	}

	var buf bytes.Buffer
	if err := nbt.Marshal(val, "", &buf); err != nil {
		t.Fatal(err)
	}

	var decoded TestStruct
	if err := nbt.Unmarshal(bufio.NewReader(&buf), &decoded); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(val, decoded) {
		t.Fatalf("struct %#v != %#v", val, decoded)
	}
}
