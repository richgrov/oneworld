package oneworld

import (
	"github.com/richgrov/oneworld/internal/level"
)

func chunkCoordsToIndex(x int32, y int32, z int32) int32 {
	return x*16*128 + z*128 + y
}

type chunk struct {
	data    level.ChunkData
	viewers map[*Player]bool
}

func newChunk() *chunk {
	return &chunk{
		viewers: make(map[*Player]bool),
	}
}

func (ch *chunk) isDataLoaded() bool {
	return ch.data.Blocks != nil
}
