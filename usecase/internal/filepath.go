package internal

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func ListRelativeFilePaths(dir, ignore string) ([]string, error) {

	filenames := []string{}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {

		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if strings.HasPrefix(path, ignore) {
			return nil
		}

		path, err = filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		filenames = append(filenames, path)
		return nil

	})

	return filenames, err
}

func ListFilepathsRecursively(paths []string, ignore string) ([]string, error) {

	files := []string{}

	ignore, err := filepath.Abs(ignore)
	if err != nil {
		return nil, err
	}

	for _, path := range paths {

		path, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}

		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}

		if info.IsDir() {

			filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {

				if err != nil {
					return err
				}

				if info.IsDir() {
					return nil
				}

				if strings.HasPrefix(path, ignore) {
					return nil
				}

				files = append(files, path)
				return nil

			})

		} else {
			files = append(files, path)
		}

	}

	return files, nil

}
