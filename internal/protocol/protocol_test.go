package protocol

import (
	"bytes"
	"testing"
	"unicode/utf16"
)

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
