package internal

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strconv"
)

func OID(data []byte) (string, error) {
	sha1 := sha1.New()
	_, err := sha1.Write(data)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(sha1.Sum(nil)), nil
}

func Pack(oid string) ([]byte, error) {

	packed := []byte{}

	for i := 0; i < len(oid); i += 2 {

		pair := oid[i : i+2]

		upper, err := strconv.ParseInt(string(pair[0]), 16, 8)
		if err != nil {
			return nil, err
		}

		lower, err := strconv.ParseInt(string(pair[1]), 16, 8)
		if err != nil {
			return nil, err
		}

		b := byte((upper << 4) + lower)

		packed = append(packed, b)
	}

	return packed, nil
}

func Unpack(bs []byte) string {
	return fmt.Sprintf("%x", bs)
}

func MustPack(oid string) []byte {

	bs, err := Pack(oid)
	if err != nil {
		panic(err)
	}

	return bs
}
