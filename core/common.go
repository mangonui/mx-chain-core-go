package core

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
)

// EmptyChannel empties the given channel
func EmptyChannel(ch chan bool) int {
	readsCnt := 0
	for {
		select {
		case <-ch:
			readsCnt++
		default:
			return readsCnt
		}
	}
}

// UniqueIdentifier returns a unique string identifier of 32 bytes
func UniqueIdentifier() string {
	return uniqueIdentifierFromReader(rand.Reader)
}

func uniqueIdentifierFromReader(reader io.Reader) string {
	buff := make([]byte, 32)
	if _, err := io.ReadFull(reader, buff); err != nil {
		panic(fmt.Errorf("cannot generate unique identifier: %w", err))
	}
	return string(buff)
}

// FileExists returns true if the file at the given path exists
func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

// GetPBFTThreshold returns the pBFT threshold for a given consensus size
func GetPBFTThreshold(consensusSize int) int {
	if consensusSize <= 0 {
		return 0
	}
	return consensusSize*2/3 + 1
}

// GetPBFTFallbackThreshold returns the pBFT fallback threshold for a given consensus size
func GetPBFTFallbackThreshold(consensusSize int) int {
	if consensusSize <= 0 {
		return 0
	}
	return consensusSize*1/2 + 1
}
