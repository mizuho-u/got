package fs

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mizuho-u/got/model"
	"github.com/mizuho-u/got/model/object"
)

type Objects struct {
	gotpath string
}

func NewObjects(gotpath string) *Objects {
	return &Objects{gotpath: gotpath}
}

func (s *Objects) Store(objects ...object.Object) error {

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

func (s *Objects) Load(oid string) (object.Object, error) {

	path := filepath.Join(s.gotpath, "objects", oid[0:2], oid[2:])
	if !s.isExist(path) {
		return nil, fmt.Errorf("%s not found", oid)
	}

	compressed, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	data, err := s.decompress(compressed)
	if err != nil {
		return nil, err
	}

	return object.ParseObject(data)
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

func (s *Objects) decompress(data []byte) ([]byte, error) {

	b := bytes.NewBuffer(data)

	zw, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(zw)
}

func (s *Objects) StoreAll(objects ...object.Object) error {

	for _, o := range objects {
		if err := s.Store(o); err != nil {
			return err
		}
	}

	return nil

}

func (s *Objects) ScanTree(oid string) model.TreeScanner {
	return newTreeScanner(s.gotpath, oid)
}

type treeScanner struct {
	gotroot  string
	rootTree string
}

func newTreeScanner(gotroot, rootTree string) *treeScanner {
	return &treeScanner{gotroot, rootTree}
}

func (ts *treeScanner) Walk(f func(name string, obj object.Entry)) {
	ts.walk(ts.rootTree, "", f)
}

func (ts *treeScanner) walk(oid, path string, f func(name string, obj object.Entry)) {

	o, err := ts.load(oid)
	if err != nil {
		return
	}

	tree, err := object.ParseTree(o)
	if err != nil {
		return
	}

	ctree := []object.Entry{}
	for _, entry := range tree.Children() {

		if entry.IsTree() {
			ctree = append(ctree, entry)
			continue
		}

		f(filepath.Join(path, entry.Basename()), entry)
	}

	for _, entry := range ctree {
		f(filepath.Join(path, entry.Basename()), entry)
		ts.walk(entry.OID(), filepath.Join(path, entry.Basename()), f)
	}

}

func (ts *treeScanner) load(oid string) (object.Object, error) {

	path := filepath.Join(ts.gotroot, "objects", oid[0:2], oid[2:])
	if !ts.isExist(path) {
		return nil, fmt.Errorf("%s not found", oid)
	}

	compressed, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	data, err := ts.decompress(compressed)
	if err != nil {
		return nil, err
	}

	return object.ParseObject(data)
}

func (ts *treeScanner) isExist(path string) bool {

	// the path exists if err is nil
	if _, err := os.Stat(path); err == nil {
		return true
	}

	return false

}

func (ts *treeScanner) decompress(data []byte) ([]byte, error) {

	b := bytes.NewBuffer(data)

	zw, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(zw)
}
