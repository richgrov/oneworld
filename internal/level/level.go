package level

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"os"
)

type LevelData struct {
	RandomSeed int64
	SpawnX     int32
	SpawnY     int32
	SpawnZ     int32
}

func ReadLevelData(levelDataFile string) (*LevelData, error) {
	file, err := os.Open(levelDataFile)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	parsed, err := readNbt(bufio.NewReader(gzipReader))
	if err != nil {
		return nil, err
	}

	data, ok := parsed["Data"].(map[string]any)
	if !ok {
		return nil, errors.New("file does not contain 'Data' tag")
	}

	// Lazy unmarshal by recoding as json
	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var levelData LevelData
	err = json.Unmarshal(encoded, &levelData)
	return &levelData, err
}
