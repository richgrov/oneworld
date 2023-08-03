package oneworld

import (
	"bytes"
	"compress/zlib"

	"github.com/richgrov/oneworld/blocks"
)

func chunkCoordsToIndex(x, y, z int) int {
	return x*16*128 + z*128 + y
}

type Chunk struct {
	x          int
	z          int
	blocks     []blocks.Block
	blockLight []byte
	skyLight   []byte
	observers  []chunkObserver
}

type chunkObserver interface {
	initializeChunk(chunkX, chunkZ int)
	unloadChunk(chunkX, chunkZ int)
	sendChunk(chunkX, chunkZ int, chunk *Chunk)
	SendBlockChange(x, y, z int, ty blocks.BlockType, data byte)
}

func (ch *Chunk) isDataLoaded() bool {
	return ch.blocks != nil
}

func (chunk *Chunk) AddObserver(observer chunkObserver) {
	chunk.observers = append(chunk.observers, observer)
	observer.initializeChunk(chunk.x, chunk.z)
	if chunk.isDataLoaded() {
		observer.sendChunk(chunk.x, chunk.z, chunk)
	}
}

func (chunk *Chunk) RemoveObserver(observer chunkObserver) {
	for i, obs := range chunk.observers {
		if obs == observer {
			observer.unloadChunk(chunk.x, chunk.z)
			chunk.observers = append(chunk.observers[:i], chunk.observers[i+1:]...)
			break
		}
	}
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
