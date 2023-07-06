package level

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/richgrov/oneworld/internal/util"
	"github.com/richgrov/oneworld/nbt"
)

const compressionZlib = 2

type McRegionLoader struct {
	WorldDir string
}

type McRegionLevelFile struct {
	Data McRegionLevelData
}

type McRegionLevelData struct {
	SpawnX     int32
	SpawnY     int32
	SpawnZ     int32
	RandomSeed int64
}

type McRegionChunkContainer struct {
	Level McRegionChunkData
}

type McRegionChunkData struct {
	Blocks     []byte
	Data       []byte
	BlockLight []byte
	SkyLight   []byte
}

func (loader *McRegionLoader) ReadWorldInfo() (WorldInfo, error) {
	info := WorldInfo{}

	file, err := os.Open(path.Join(loader.WorldDir, "level.dat"))
	defer file.Close()
	if err != nil {
		return info, err
	}

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return info, err
	}

	var level McRegionLevelFile
	if err := nbt.Unmarshal(bufio.NewReader(gzipReader), &level); err != nil {
		return info, err
	}

	info.SpawnX = level.Data.SpawnX
	info.SpawnY = level.Data.SpawnY
	info.SpawnZ = level.Data.SpawnZ
	info.BiomeSeed = level.Data.RandomSeed
	return info, err
}

func (loader *McRegionLoader) LoadChunks(chunks []ChunkPos, consumer chan ChunkReadResult) {
	files := make(map[ChunkPos]*os.File)
	defer func() {
		for _, file := range files {
			file.Close()
		}
	}()

	for _, chunkPos := range chunks {
		result := ChunkReadResult{
			Pos: ChunkPos(chunkPos),
		}

		region := ChunkPos{
			util.DivideAndFloorI32(chunkPos.X, 32),
			util.DivideAndFloorI32(chunkPos.Z, 32),
		}

		file, ok := files[region]
		if !ok {
			regionFile, err := os.Open(filepath.Join(loader.WorldDir, "region", fmt.Sprintf("r.%d.%d.mcr", region.X, region.Z)))
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					// File doesn't exist- chunk simply isn't generated
					result.Data = &ChunkData{}
				} else {
					result.Error = err
				}
				consumer <- result
				continue
			}

			file = regionFile
			files[region] = regionFile
		}

		result.Data, result.Error = readChunk(file, chunkPos)
		consumer <- result
	}
}

func readChunk(file *os.File, pos ChunkPos) (*ChunkData, error) {
	offset, err := getChunkPositionInFile(file, pos)
	if err != nil {
		return nil, err
	}

	// 0 offset means the chunk is not present in the region
	if offset == 0 {
		return &ChunkData{}, nil
	}

	_, err = file.Seek(int64(offset*1024*4 /* offset is in 4KiB sectors */), 0)
	if err != nil {
		return nil, err
	}

	var dataLength int32
	if err := binary.Read(file, binary.BigEndian, &dataLength); err != nil {
		return nil, err
	}

	data := make([]byte, dataLength)
	if _, err := file.Read(data); err != nil {
		return nil, err
	}

	return parseChunkData(data)
}

func getChunkPositionInFile(file *os.File, pos ChunkPos) (int32, error) {
	offsetPos := 4 * ((pos.X & 31) + (pos.Z&31)*32)
	_, err := file.Seek(int64(offsetPos), 0)
	if err != nil {
		return 0, err
	}

	var offset int32
	if err := binary.Read(file, binary.BigEndian, &offset); err != nil {
		return 0, err
	}

	if offset == 0 {
		return 0, nil
	}

	return offset >> 8, nil
}

func parseChunkData(data []byte) (*ChunkData, error) {
	if data[0] != compressionZlib {
		return nil, errors.New("unsupported compression type")
	}

	reader, err := zlib.NewReader(bytes.NewBuffer(data[1:]))
	if err != nil {
		return nil, err
	}

	var chunk McRegionChunkContainer
	if err := nbt.Unmarshal(bufio.NewReader(reader), &chunk); err != nil {
		return nil, err
	}

	return &ChunkData{
		Blocks:     chunk.Level.Blocks,
		BlockData:  nibblesToBytes(chunk.Level.Data),
		BlockLight: nibblesToBytes(chunk.Level.BlockLight),
		SkyLight:   nibblesToBytes(chunk.Level.SkyLight),
	}, nil
}

func nibblesToBytes(nibbles []byte) []byte {
	buf := make([]byte, len(nibbles)*2)
	for i, b := range nibbles {
		buf[i*2] = b & 0b00001111
		buf[i*2+1] = b >> 4
	}
	return buf
}
