package level

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

type ChunkPos struct {
	X int32
	Z int32
}

type ChunkData struct {
	Blocks     []byte
	BlockData  []byte
	BlockLight []byte
	SkyLight   []byte
}

// Compress chunk data to the format it is sent over Minecraft protocol
func (chunk *ChunkData) CompressData() []byte {
	data := bytes.NewBuffer(make([]byte, 0))
	data.Write(chunk.Blocks)
	data.Write(chunk.BlockData)
	data.Write(chunk.BlockLight)
	data.Write(chunk.SkyLight)

	var out bytes.Buffer
	w := zlib.NewWriter(&out)
	w.Write(data.Bytes())
	w.Close()

	return out.Bytes()
}

// Tries to load all chunks at the specified positions. The chunks are returned
// in the same order they were specified
func LoadChunks(regionDir string, chunks []ChunkPos) []*ChunkData {
	files := make(map[ChunkPos]*os.File)
	defer func() {
		for _, file := range files {
			file.Close()
		}
	}()

	results := make([]*ChunkData, 0, len(chunks))

	for i, chunkPos := range chunks {
		region := ChunkPos{chunkPos.X / 32, chunkPos.Z / 32}
		results = append(results, &ChunkData{})

		// Get cached file handle or open a new one
		file, ok := files[region]
		if !ok {
			regionFile, err := os.Open(filepath.Join(regionDir, fmt.Sprintf("r.%d.%d.mcr", region.X, region.Z)))
			if err != nil {
				continue
			}
			file = regionFile
			files[region] = regionFile
		}

		offsetPos := 4 * ((chunkPos.X & 31) + (chunkPos.Z&31)*32)
		_, err := file.Seek(int64(offsetPos), 0)
		if err != nil {
			fmt.Printf("%s\n", err)
			continue
		}

		var offset int32
		if err := binary.Read(file, binary.BigEndian, &offset); err != nil {
			fmt.Printf("%s\n", err)
			continue
		}

		chunkOffset := offset >> 8
		chunkLength := byte(offset)
		if chunkLength == 0 {
			// Chunk not present in region
			continue
		}

		_, err = file.Seek(int64(chunkOffset*1024*4), 0)
		if err != nil {
			fmt.Printf("%s\n", err)
			continue
		}

		var exactLength int32
		if err := binary.Read(file, binary.BigEndian, &exactLength); err != nil {
			fmt.Printf("%s\n", err)
			continue
		}

		if exactLength < 1 {
			continue
		}

		data := make([]byte, exactLength)
		if _, err := file.Read(data); err != nil {
			fmt.Printf("%s\n", err)
			continue
		}

		if data[0] != 2 {
			println("invalid compression type")
			continue
		}

		uncompressor, err := zlib.NewReader(bytes.NewBuffer(data[1:]))
		if err != nil {
			fmt.Printf("%s\n", err)
			continue
		}

		nbt, err := readNbt(bufio.NewReader(uncompressor))
		if err != nil {
			fmt.Printf("%s\n", err)
			continue
		}

		level, ok := nbt["Level"].(map[string]any)
		if !ok {
			fmt.Printf("chunk does not contain 'Level' tag\n")
			continue
		}

		blocks, ok := level["Blocks"].([]byte)
		if !ok {
			fmt.Printf("chunk does not contain 'Blocks' tag\n")
			continue
		}

		blockData, ok := level["Data"].([]byte)
		if !ok {
			fmt.Printf("chunk does not contain 'Data' tag\n")
		}

		blockLight, ok := level["BlockLight"].([]byte)
		if !ok {
			fmt.Printf("chunk does not contain 'BlockLight' tag\n")
		}

		skyLight, ok := level["SkyLight"].([]byte)
		if !ok {
			fmt.Printf("chunk does not contain 'SkyLight' tag\n")
		}

		chunkData := results[i]
		chunkData.Blocks = blocks
		chunkData.BlockData = blockData
		chunkData.BlockLight = blockLight
		chunkData.SkyLight = skyLight
	}

	return results
}
