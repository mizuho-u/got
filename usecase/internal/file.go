package internal

import (
	"os"
	"path/filepath"
)

func ReadFile(path string) ([]byte, error) {

	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return data, nil

}
