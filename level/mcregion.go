package level

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/richgrov/oneworld/internal/util"
)

const compressionZlib = 2

type McRegionLoader struct {
	WorldDir string
}

type McRegionLevelData struct {
	SpawnX     int32
	SpawnY     int32
	SpawnZ     int32
	RandomSeed int64
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
	parsed, err := readNbt(bufio.NewReader(gzipReader))
	if err != nil {
		return info, err
	}

	data, ok := parsed["Data"].(map[string]any)
	if !ok {
		return info, errors.New("file does not contain 'Data' tag")
	}

	// Lazy unmarshal by recoding as json
	encoded, err := json.Marshal(data)
	if err != nil {
		return info, err
	}

	var level McRegionLevelData
	err = json.Unmarshal(encoded, &level)
	info.SpawnX = level.SpawnX
	info.SpawnY = level.SpawnY
	info.SpawnZ = level.SpawnZ
	info.BiomeSeed = level.RandomSeed

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

	uncompressor, err := decompressChunkData(data)
	if err != nil {
		return nil, err
	}

	return readChunkNbt(uncompressor)
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

func decompressChunkData(data []byte) (io.Reader, error) {
	if data[0] != compressionZlib {
		return nil, errors.New("unsupported compression type")
	}

	return zlib.NewReader(bytes.NewBuffer(data[1:]))
}

func readChunkNbt(r io.Reader) (*ChunkData, error) {
	nbt, err := readNbt(bufio.NewReader(r))
	if err != nil {
		return nil, err
	}

	level, ok := nbt["Level"].(map[string]any)
	if !ok {
		return nil, errors.New("missing 'Level' tag")
	}

	blocks, ok := level["Blocks"].([]byte)
	if !ok {
		return nil, errors.New("missing 'Blocks' tag")
	}

	blockData, ok := level["Data"].([]byte)
	if !ok {
		return nil, errors.New("missing 'Data' tag")
	}

	blockLight, ok := level["BlockLight"].([]byte)
	if !ok {
		return nil, errors.New("missing 'BlockLight' tag")
	}

	skyLight, ok := level["SkyLight"].([]byte)
	if !ok {
		return nil, errors.New("missing 'SkyLight' tag")
	}

	return &ChunkData{
		Blocks:     blocks,
		BlockData:  nibblesToBytes(blockData),
		BlockLight: nibblesToBytes(blockLight),
		SkyLight:   nibblesToBytes(skyLight),
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
