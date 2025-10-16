package util

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
)

// FileHasher defines the interface for calculating file hashes
type FileHasher interface {
	CalculateHash(reader io.Reader) (string, error)
}

// SHA256Hasher implements FileHasher using SHA256 algorithm
type SHA256Hasher struct{}

// NewSHA256Hasher creates a new SHA256 file hasher
func NewSHA256Hasher() FileHasher {
	return &SHA256Hasher{}
}

// CalculateHash computes the SHA256 hash of the data from the reader
func (h *SHA256Hasher) CalculateHash(reader io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// GenericHasher implements FileHasher using a custom hash function
type GenericHasher struct {
	hashFunc func() hash.Hash
}

// NewGenericHasher creates a new generic file hasher with a custom hash function
func NewGenericHasher(hashFunc func() hash.Hash) FileHasher {
	return &GenericHasher{hashFunc: hashFunc}
}

// CalculateHash computes the hash of the data from the reader using the configured hash function
func (h *GenericHasher) CalculateHash(reader io.Reader) (string, error) {
	hasher := h.hashFunc()
	if _, err := io.Copy(hasher, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
