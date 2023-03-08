package database

import (
	"bytes"
	"compress/zlib"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mizuho-u/got/model/object"
)

type Objects struct {
	gotpath string
}

func NewObjects(gotpath string) *Objects {
	return &Objects{gotpath: gotpath}
}

func (s *Objects) Store(o object.Object) error {

	path := filepath.Join(s.gotpath, "objects", o.OID()[0:2], o.OID()[2:])
	if s.isExist(path) {
		return nil
	}

	compressed, err := s.compress(o.Content())
	if err != nil {
		return err
	}

	return s.create(path, compressed)
}

func (s *Objects) isExist(path string) bool {

	// the path exists if err is nil
	if _, err := os.Stat(path); err == nil {
		return true
	}

	return false

}

func (s *Objects) create(path string, data []byte) error {

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	temp, err := ioutil.TempFile(filepath.Dir(path), "tmp_obj_*")
	if err != nil {
		return err
	}

	if _, err := temp.Write(data); err != nil {
		return err
	}

	return os.Rename(temp.Name(), path)
}

func (s *Objects) compress(data []byte) ([]byte, error) {

	var b bytes.Buffer

	zw, err := zlib.NewWriterLevel(&b, zlib.BestSpeed)
	if err != nil {
		return nil, err
	}
	defer zw.Close()

	_, err = zw.Write(data)
	if err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (s *Objects) StoreAll(objects ...object.Object) error {

	for _, o := range objects {
		if err := s.Store(o); err != nil {
			return err
		}
	}

	return nil

}
