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
	Blocks     [ChunkSize]blocks.Block
	BlockLight [ChunkSize]byte
	SkyLight   [ChunkSize]byte
}

type ChunkPos struct {
	X int
	Z int
}

func (chunk *Chunk) Set(x, y, z int, block blocks.Block) {
	index := chunkCoordsToIndex(x, y, z)
	chunk.Blocks[index] = block
}

func (chunk *Chunk) SetBlockLight(x, y, z int, level byte) {
	index := chunkCoordsToIndex(x, y, z)
	chunk.BlockLight[index] = level
}

func (chunk *Chunk) SetSkyLight(x, y, z int, level byte) {
	index := chunkCoordsToIndex(x, y, z)
	chunk.SkyLight[index] = level
}

func (ch *Chunk) serializeToNetwork() []byte {
	capacity := 16*16*128 + 16*16*64 + 16*16*64 + 16*16*64
	data := bytes.NewBuffer(make([]byte, 0, capacity))

	for _, block := range ch.Blocks {
		data.WriteByte(byte(block.Type))
	}

	for i := 0; i < len(ch.Blocks); i += 2 {
		data.WriteByte(byte(ch.Blocks[i].Data)&0b00001111 | byte(ch.Blocks[i+1].Data)<<4)
	}

	packToNibbleArray(ch.BlockLight[:], data)
	packToNibbleArray(ch.SkyLight[:], data)

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
