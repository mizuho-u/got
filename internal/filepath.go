package internal

import (
	"path/filepath"
	"strings"
)

func ParentDirs(path string) []string {

	parentsDirs := []string{}
	dir := filepath.Dir(path)
	if dir == "." {
		return []string{}
	}

	dirs := strings.Split(filepath.Dir(path), "/")
	for i := 1; i <= len(dirs); i++ {
		parentsDirs = append(parentsDirs, filepath.Join(dirs[0:i]...))
	}

	return parentsDirs
}
