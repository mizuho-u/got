package internal

import (
	"encoding/binary"
)

func UintToBytes[T uint16 | uint32 | uint64](v T) []byte {

	var bs []byte

	switch n := any(v).(type) {
	case uint16:
		bs = make([]byte, 2)
		binary.BigEndian.PutUint16(bs, n)
	case uint32:
		bs = make([]byte, 4)
		binary.BigEndian.PutUint32(bs, n)
	case uint64:
		bs = make([]byte, 8)
		binary.BigEndian.PutUint64(bs, n)
	}

	return bs
}

func BytesToUint32(bs []byte) uint32 {

	return binary.BigEndian.Uint32(bs)

}
