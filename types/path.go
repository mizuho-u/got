package types

import "path/filepath"

type Path string

type Basename Path

type AbsPath Path

func (abs AbsPath) Basename() Basename {
	return Basename(filepath.Base(string(abs)))
}

func (abs AbsPath) Rel(basepath AbsPath) (RelPath, error) {

	rel, err := filepath.Rel(string(basepath), string(abs))
	if err != nil {
		return "", nil
	}

	return RelPath(string(rel)), nil
}

func (abs AbsPath) Join(element ...Path) AbsPath {

	path := string(abs)
	for _, e := range element {
		path = filepath.Join(path, string(e))
	}

	return AbsPath(path)

}

type RelPath Path

func (abs RelPath) Join(element ...Path) RelPath {

	path := string(abs)
	for _, e := range element {
		path = filepath.Join(path, string(e))
	}

	return RelPath(path)

}

func (rel RelPath) Abs() (AbsPath, error) {

	abs, err := filepath.Abs(string(rel))

	return AbsPath(abs), err

}

// func join[T ~string](element ...Path) T {

// 	path := ""
// 	for _, e := range element {
// 		path = filepath.Join(path, string(e))
// 	}

// 	return T("aaa")

// }
