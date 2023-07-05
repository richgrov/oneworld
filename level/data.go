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
	X int32
	Z int32
}

// The length of each slice in this struct should be equal to `ChunkSize`. None
// of the slices should be nil
type ChunkData struct {
	Blocks     []byte
	BlockData  []byte
	BlockLight []byte
	SkyLight   []byte
}

func (cd *ChunkData) InitializeToAir() {
	cd.Blocks = make([]byte, ChunkSize)
	cd.BlockData = make([]byte, ChunkSize)
	cd.BlockLight = make([]byte, ChunkSize)
	cd.SkyLight = make([]byte, ChunkSize)
}

func (cd *ChunkData) Set(x uint, y uint, z uint, ty blocks.BlockType, data byte) {
	index := chunkPosToIndex(x, y, z)
	cd.Blocks[index] = byte(ty)
	cd.BlockData[index] = data
}

type ChunkReadResult struct {
	Pos   ChunkPos
	Data  *ChunkData
	Error error
}
