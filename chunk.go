package oneworld

import (
	"github.com/richgrov/oneworld/internal/level"
)

type chunk struct {
	data    level.ChunkData
	viewers []*Player
}

func newChunk() *chunk {
	return &chunk{
		viewers: make([]*Player, 0),
	}
}

func (ch *chunk) isDataLoaded() bool {
	return ch.data.Blocks != nil
}
