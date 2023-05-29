package oneworld

import (
	"github.com/richgrov/oneworld/internal/level"
)

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
