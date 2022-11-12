package apparmor

import (
	"crypto/rand"
	"fmt"
)

func GetMagicToken() (uint64, error) {
	var buf [64 / 8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return 0, fmt.Errorf("failed to generate magic token: %v", err)
	}
	return uint64(buf[0])<<56 | uint64(buf[1])<<48 |
		uint64(buf[2])<<40 | uint64(buf[3])<<32 |
		uint64(buf[4])<<24 | uint64(buf[5])<<16 |
		uint64(buf[6])<<8 | uint64(buf[7]), nil
}
