package level

import "github.com/richgrov/oneworld/blocks"

const ChunkSize = 16 * 16 * 128

func chunkPosToIndex(x uint, y uint, z uint) uint {
	return x*16*128 + z*128 + y
}

type WorldInfo struct {
	BiomeSeed int64
	SpawnX    int32
	SpawnY    int32
	SpawnZ    int32
}

type ChunkPos struct {
	X int
	Z int
}

// The length of each slice in this struct should be equal to `ChunkSize`. None
// of the slices should be nil
type ChunkData struct {
	Blocks     []blocks.Block
	BlockLight []byte
	SkyLight   []byte
}

func (cd *ChunkData) InitializeToAir() {
	cd.Blocks = make([]blocks.Block, ChunkSize)
	cd.BlockLight = make([]byte, ChunkSize)
	cd.SkyLight = make([]byte, ChunkSize)
}

func (cd *ChunkData) Set(x, y, z uint, block blocks.Block) {
	index := chunkPosToIndex(x, y, z)
	cd.Blocks[index] = block
}

type ChunkReadResult struct {
	ChunkX int
	ChunkZ int
	Data   *ChunkData
	Error  error
}
