package model

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"sort"

	"github.com/mizuho-u/got/model/internal"
)

const (
	headerSignature string = "DIRC"
	headerVersion   uint32 = 2
)

type Index interface {
	Serialize() ([]byte, error)
}

type index struct {
	entries map[string]*indexEntry
}

type indexOption func(*index) error

func indexSource(data io.Reader) indexOption {

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

		index.add(entries...)

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

func newIndex(opts ...indexOption) (*index, error) {

	i := &index{entries: map[string]*indexEntry{}}

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

func parseEntries(data io.Reader, count uint32) (raw []byte, entries []*indexEntry, err error) {

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

func (i *index) add(entries ...*indexEntry) {

	for _, entry := range entries {
		i.entries[entry.filename] = entry
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

type indexEntry struct {
	filename string
	oid      string
	stat     *FileStat
}

func NewIndexEntry(name, oid string, stat *FileStat) *indexEntry {

	return &indexEntry{
		filename: name,
		oid:      oid,
		stat:     stat,
	}

}

func parseIndexEntry(entry []byte) *indexEntry {

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

func (ie *indexEntry) serialize() []byte {

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
