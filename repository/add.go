package repository

import (
	"io"

	"github.com/mizuho-u/got/repository/object"
)

func (repo *repository) Add(scanner WorkspaceScanner) (objects []object.Object, err error) {

	for {

		f, err := scanner.Next()
		if err != nil {
			return nil, err
		}
		if f == nil {
			return objects, nil
		}

		data, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}

		blob, err := object.NewBlob(f.Name(), data)
		if err != nil {
			return nil, err
		}
		objects = append(objects, blob)

		repo.index.Add(NewIndexEntry(f.Name(), blob.OID(), f.Stats()))

	}

}
