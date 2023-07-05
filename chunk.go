package oneworld

import (
	"bytes"
	"compress/zlib"

	"github.com/richgrov/oneworld/blocks"
	"github.com/richgrov/oneworld/level"
)

func chunkCoordsToIndex(x int32, y int32, z int32) int32 {
	return x*16*128 + z*128 + y
}

type chunk struct {
	blocks     []Block
	blockLight []byte
	skyLight   []byte
	viewers    map[*Player]bool
}

func newChunk() *chunk {
	return &chunk{
		viewers: make(map[*Player]bool),
	}
}

func (ch *chunk) isDataLoaded() bool {
	return ch.blocks != nil
}

func (ch *chunk) initialize(data *level.ChunkData) {
	ch.blocks = make([]Block, 16*16*128)
	for i := 0; i < len(ch.blocks); i++ {
		ch.blocks[i].ty = blocks.BlockType(data.Blocks[i])
		ch.blocks[i].data = data.BlockData[i]
	}
	ch.blockLight = data.BlockLight
	ch.skyLight = data.SkyLight
}

func (ch *chunk) serializeToNetwork() []byte {
	capacity := 16*16*128 + 16*16*64 + 16*16*64 + 16*16*64
	data := bytes.NewBuffer(make([]byte, 0, capacity))

	for _, block := range ch.blocks {
		data.WriteByte(byte(block.Type()))
	}

	for i := 0; i < len(ch.blocks); i += 2 {
		data.WriteByte(ch.blocks[i].Data()&0b00001111 | ch.blocks[i+1].Data()<<4)
	}

	packToNibbleArray(ch.blockLight, data)
	packToNibbleArray(ch.skyLight, data)

	var out bytes.Buffer
	w := zlib.NewWriter(&out)
	w.Write(data.Bytes())
	w.Close()

	return out.Bytes()
}

func packToNibbleArray(buf []byte, out *bytes.Buffer) {
	for i := 0; i < len(buf); i += 2 {
		out.WriteByte(buf[i]&0b00001111 | buf[i+1]<<4)
	}
}
