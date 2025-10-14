package util

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
)

type FileHasher interface {
	CalculateHash(reader io.Reader) (string, error)
}

type SHA256Hasher struct{}

func NewSHA256Hasher() FileHasher {
	return &SHA256Hasher{}
}

func (h *SHA256Hasher) CalculateHash(reader io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

type GenericHasher struct {
	hashFunc func() hash.Hash
}

func NewGenericHasher(hashFunc func() hash.Hash) FileHasher {
	return &GenericHasher{hashFunc: hashFunc}
}

func (h *GenericHasher) CalculateHash(reader io.Reader) (string, error) {
	hasher := h.hashFunc()
	if _, err := io.Copy(hasher, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
