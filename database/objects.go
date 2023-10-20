package database

import (
	"bytes"
	"compress/zlib"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mizuho-u/got/model/object"
)

type objects struct {
	gotpath string
}

func NewObjects(gotpath string) *objects {
	return &objects{gotpath: gotpath}
}

func (s *objects) Store(objects ...object.Object) error {

	for _, o := range objects {

		path := filepath.Join(s.gotpath, "objects", o.OID()[0:2], o.OID()[2:])
		if s.isExist(path) {
			continue
		}

		compressed, err := s.compress(o.Content())
		if err != nil {
			return err
		}

		if err := s.create(path, compressed); err != nil {
			return err
		}

	}

	return nil

}

func (s *objects) isExist(path string) bool {

	// the path exists if err is nil
	if _, err := os.Stat(path); err == nil {
		return true
	}

	return false

}

func (s *objects) create(path string, data []byte) error {

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

func (s *objects) compress(data []byte) ([]byte, error) {

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

func (s *objects) StoreAll(objects ...object.Object) error {

	for _, o := range objects {
		if err := s.Store(o); err != nil {
			return err
		}
	}

	return nil

}
