package internal

import (
	"crypto/sha1"
	"encoding/hex"
)

func VerifySHA1(p []byte, expected string) bool {

	sha1 := sha1.New()
	if _, err := sha1.Write(p); err != nil {
		return false
	}

	return hex.EncodeToString(sha1.Sum(nil)) == expected
}

type Checksum struct {
	p []byte
}

func NewChecksum() *Checksum {
	return &Checksum{}
}

func (c *Checksum) Write(p []byte) {
	c.p = append(c.p, p...)
}

func (c *Checksum) Expect(s string) bool {

	sha1 := sha1.New()
	if _, err := sha1.Write(c.p); err != nil {
		return false
	}

	return hex.EncodeToString(sha1.Sum(nil)) == s

}
