package oneworld

import (
	"bytes"
	"compress/zlib"

	"github.com/richgrov/oneworld/blocks"
)

const ChunkSize = 16 * 16 * 128

func chunkCoordsToIndex(x, y, z int) int {
	return x*16*128 + z*128 + y
}

type Chunk struct {
	blocks     [ChunkSize]blocks.Block
	blockLight [ChunkSize]byte
	skyLight   [ChunkSize]byte
}

type ChunkPos struct {
	X int
	Z int
}

func (chunk *Chunk) Set(x, y, z int, block blocks.Block) {
	index := chunkCoordsToIndex(x, y, z)
	chunk.blocks[index] = block
}

func (chunk *Chunk) SetBlockLight(x, y, z int, level byte) {
	index := chunkCoordsToIndex(x, y, z)
	chunk.blockLight[index] = level
}

func (chunk *Chunk) SetSkyLight(x, y, z int, level byte) {
	index := chunkCoordsToIndex(x, y, z)
	chunk.skyLight[index] = level
}

func (ch *Chunk) serializeToNetwork() []byte {
	capacity := 16*16*128 + 16*16*64 + 16*16*64 + 16*16*64
	data := bytes.NewBuffer(make([]byte, 0, capacity))

	for _, block := range ch.blocks {
		data.WriteByte(byte(block.Type))
	}

	for i := 0; i < len(ch.blocks); i += 2 {
		data.WriteByte(ch.blocks[i].Data&0b00001111 | ch.blocks[i+1].Data<<4)
	}

	packToNibbleArray(ch.blockLight[:], data)
	packToNibbleArray(ch.skyLight[:], data)

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
