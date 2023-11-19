package repository

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"sort"

	"github.com/mizuho-u/got/repository/internal"
	"github.com/mizuho-u/got/repository/object"
)

const (
	headerSignature string = "DIRC"
	headerVersion   uint32 = 2
)

type Index interface {
	IndexSerializable
	IndexWriter
	tracked(name string) bool
	match(e WorkspaceEntry) bool
}

type IndexSerializable interface {
	Serialize() ([]byte, error)
}

type IndexWriter interface {
	Add(entries ...*IndexEntry)
	Delete(entry string)
}

type IndexReader interface {
}

type index struct {
	entries map[string]*IndexEntry
	parents map[string]map[string]struct{}
}

type indexOption func(*index) error

func IndexSource(data io.Reader) indexOption {

	return func(index *index) error {

		checksum := internal.NewChecksum()

		rawHeader, _, _, count, err := parseHeader(data)
		if err != nil {
			return err
		}
		checksum.Write(rawHeader)

		rawEntries, entries, err := parseEntries(data, count)
		if err != nil {
			return err
		}
		checksum.Write(rawEntries)

		index.Add(entries...)

		digest, err := io.ReadAll(data)
		if err != nil {
			return err
		}

		if !checksum.Expect(internal.Unpack(digest)) {
			return errors.New("index check sum not match")
		}

		return err

	}

}

func NewIndex(opts ...indexOption) (*index, error) {

	i := &index{entries: map[string]*IndexEntry{}, parents: map[string]map[string]struct{}{}}

	for _, opt := range opts {
		if err := opt(i); err != nil {
			return nil, err
		}
	}

	return i, nil
}

func read(reader io.Reader, count int) ([]byte, error) {

	bs := make([]byte, count)

	_, err := reader.Read(bs)
	if err != nil {
		return nil, err
	}

	return bs, nil

}

func parseHeader(data io.Reader) (raw []byte, signature string, version uint32, count uint32, err error) {

	raw, err = read(data, 12)
	if err != nil {
		return
	}

	signature = string(raw[0:4])
	if string(raw[0:4]) != headerSignature {
		err = errors.New("invalid signature")
		return
	}

	version = binary.BigEndian.Uint32(raw[4:8])
	if version != 2 {
		err = errors.New("invalid version")
		return
	}

	count = binary.BigEndian.Uint32(raw[8:12])

	return
}

func parseEntries(data io.Reader, count uint32) (raw []byte, entries []*IndexEntry, err error) {

	for i := uint32(0); i < count; i++ {

		entry, e := readOneEntry(data)
		if e != nil {
			err = e
			return
		}

		entries = append(entries, parseIndexEntry(entry))

		raw = append(raw, entry...)

	}

	return
}

func readOneEntry(data io.Reader) ([]byte, error) {

	entry, err := read(data, 64)
	if err != nil {
		return nil, err
	}

	for entry[len(entry)-1] != 0x00 {

		block, err := read(data, 8)
		if err != nil {
			return nil, err
		}
		entry = append(entry, block...)
	}

	return entry, nil

}

func (i *index) Add(entries ...*IndexEntry) {

	for _, entry := range entries {
		i.discardConflicts(entry)
		i.storeEntry(entry)
	}

}

func (i *index) discardConflicts(e *IndexEntry) {

	// replacing a file with a directory
	for _, parentDir := range internal.ParentDirs(e.filename) {
		delete(i.entries, parentDir)
	}

	// replacing a directory with a file
	if fs, ok := i.parents[e.filename]; ok {

		for _, f := range internal.Keys(fs) {
			i.deleteEntry(f)
		}

	}
}

func (i *index) storeEntry(e *IndexEntry) {

	i.entries[e.filename] = e
	i.storeParent(e.filename)

}

func (i *index) Delete(entry string) {

	i.deleteEntry(entry)
	i.deleteChildren(entry)

}

func (i *index) deleteEntry(filename string) {

	delete(i.entries, filename)
	i.deleteParent(filename)

}

func (i *index) storeParent(filename string) {

	for _, p := range internal.ParentDirs(filename) {

		if _, ok := i.parents[p]; !ok {
			i.parents[p] = map[string]struct{}{}
		}

		i.parents[p][filename] = struct{}{}
	}
}

func (i *index) deleteParent(filename string) {

	for _, p := range internal.ParentDirs(filename) {
		delete(i.parents[p], filename)
	}

}

func (i *index) deleteChildren(parent string) {

	p, ok := i.parents[parent]
	if !ok {
		return
	}

	for child := range p {
		i.deleteEntry(child)
	}

}

func (i *index) Serialize() ([]byte, error) {

	content := []byte{}

	content = append(content, []byte(headerSignature)...)
	content = append(content, internal.UintToBytes(headerVersion)...)
	content = append(content, internal.UintToBytes(uint32(len(i.entries)))...)

	keys := internal.Keys(i.entries)
	sort.Strings(keys)
	for _, k := range keys {
		content = append(content, i.entries[k].serialize()...)
	}

	oid, err := internal.OID(content)
	if err != nil {
		return nil, err
	}

	packed, err := internal.Pack(oid)
	if err != nil {
		return nil, err
	}

	content = append(content, packed...)

	return content, nil
}

func (i *index) tracked(name string) bool {

	_, inEntries := i.entries[name]
	_, inParents := i.parents[name]

	return inEntries || inParents
}

func (i *index) trackedFile(name string) bool {

	_, inEntries := i.entries[name]

	return inEntries
}

func (i *index) match(e WorkspaceEntry) bool {

	entry, inEntries := i.entries[e.Name()]
	if !inEntries {
		return false
	}

	if entry.stat.size != uint32(e.Size()) {
		return false
	}

	if entry.stat.mode != e.Stats().mode {
		return false
	}

	if entry.matchTimes(e.Stats()) {
		return true
	}

	data, err := io.ReadAll(e)
	if err != nil {
		return false
	}

	blod, err := object.NewBlob(e.Name(), data)
	if err != nil {
		return false
	}

	if entry.oid != blod.OID() {
		return false
	}

	entry.stat = e.Stats()

	return true

}

type IndexEntry struct {
	filename string
	oid      string
	stat     *FileStat
}

func NewIndexEntry(name, oid string, stat *FileStat) *IndexEntry {

	return &IndexEntry{
		filename: name,
		oid:      oid,
		stat:     stat,
	}

}

func parseIndexEntry(entry []byte) *IndexEntry {

	fstat := parseFileStat(entry[0:40])

	oid := internal.Unpack(entry[40:60])

	// pathlen := entry[60:62]
	filename := string(bytes.TrimRightFunc(entry[62:], func(r rune) bool {
		return r == padding
	}))

	return NewIndexEntry(filename, oid, fstat)
}

func parseFileStat(stat []byte) *FileStat {

	ctime := internal.BytesToUint32(stat[0:4])
	ctime_nsec := internal.BytesToUint32(stat[4:8])
	mtime := internal.BytesToUint32(stat[8:12])
	mtime_nsec := internal.BytesToUint32(stat[12:16])
	dev := internal.BytesToUint32(stat[16:20])
	ino := internal.BytesToUint32(stat[20:24])
	mode := internal.BytesToUint32(stat[24:28])
	uid := internal.BytesToUint32(stat[28:32])
	gid := internal.BytesToUint32(stat[32:36])
	size := internal.BytesToUint32(stat[36:40])

	fstat := &FileStat{
		ctime:      ctime,
		ctime_nsec: ctime_nsec,
		mtime:      mtime,
		mtime_nsec: mtime_nsec,
		dev:        dev,
		ino:        ino,
		mode:       mode,
		uid:        uid,
		gid:        gid,
		size:       size,
	}

	return fstat

}

const block = 8
const padding = 0x00
const max_path_size = 4095

func (ie *IndexEntry) serialize() []byte {

	content := []byte{}

	content = append(content, internal.UintToBytes(ie.stat.ctime)...)
	content = append(content, internal.UintToBytes(ie.stat.ctime_nsec)...)
	content = append(content, internal.UintToBytes(ie.stat.mtime)...)
	content = append(content, internal.UintToBytes(ie.stat.mtime_nsec)...)
	content = append(content, internal.UintToBytes(ie.stat.dev)...)
	content = append(content, internal.UintToBytes(ie.stat.ino)...)
	content = append(content, internal.UintToBytes(ie.stat.mode)...)
	content = append(content, internal.UintToBytes(ie.stat.uid)...)
	content = append(content, internal.UintToBytes(ie.stat.gid)...)
	content = append(content, internal.UintToBytes(ie.stat.size)...)

	content = append(content, internal.MustPack(ie.oid)...)

	pathlen := internal.Min(len(ie.filename), max_path_size)
	content = append(content, internal.UintToBytes(uint16(pathlen))...)
	content = append(content, []byte(ie.filename)...)
	content = append(content, 0x00)

	for {
		if len(content)%block == 0 {
			break
		}

		content = append(content, padding)
	}

	return content

}

func (ie *IndexEntry) permission() object.Permission {
	return ie.stat.permission()
}

func (ie *IndexEntry) matchTimes(stat *FileStat) bool {

	return ie.stat.ctime == stat.ctime &&
		ie.stat.ctime_nsec == stat.ctime_nsec &&
		ie.stat.mtime == stat.mtime &&
		ie.stat.mtime_nsec == stat.mtime_nsec

}

func (ie *IndexEntry) Name() string {
	return ie.filename
}
