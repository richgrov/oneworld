package level

const ChunkSize = 16 * 16 * 128

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

type ChunkReadResult struct {
	Pos   ChunkPos
	Data  *ChunkData
	Error error
}
