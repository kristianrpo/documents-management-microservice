package util_test

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"errors"
	"testing"

	"github.com/kristianrpo/document-management-microservice/internal/application/util"
	"github.com/stretchr/testify/assert"
)

func TestSHA256Hasher_CalculateHash(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		hasher := util.NewSHA256Hasher()
		data := []byte("test data")
		reader := bytes.NewReader(data)

		hash, err := hasher.CalculateHash(reader)

		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 64) // SHA256 produces 64 hex characters

		// Verify the hash is deterministic
		reader2 := bytes.NewReader(data)
		hash2, err2 := hasher.CalculateHash(reader2)
		assert.NoError(t, err2)
		assert.Equal(t, hash, hash2)
	})

	t.Run("empty data", func(t *testing.T) {
		hasher := util.NewSHA256Hasher()
		reader := bytes.NewReader([]byte{})

		hash, err := hasher.CalculateHash(reader)

		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
		// SHA256 of empty data should be a specific hash
		assert.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", hash)
	})

	t.Run("large data", func(t *testing.T) {
		hasher := util.NewSHA256Hasher()
		// Create 1MB of data
		data := make([]byte, 1024*1024)
		for i := range data {
			data[i] = byte(i % 256)
		}
		reader := bytes.NewReader(data)

		hash, err := hasher.CalculateHash(reader)

		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 64)
	})

	t.Run("reader error", func(t *testing.T) {
		hasher := util.NewSHA256Hasher()
		reader := &errorReader{err: errors.New("read failed")}

		hash, err := hasher.CalculateHash(reader)

		assert.Error(t, err)
		assert.Empty(t, hash)
		assert.Contains(t, err.Error(), "read failed")
	})
}

func TestGenericHasher_CalculateHash(t *testing.T) {
	t.Run("with SHA256", func(t *testing.T) {
		hasher := util.NewGenericHasher(sha256.New)
		data := []byte("test data")
		reader := bytes.NewReader(data)

		hash, err := hasher.CalculateHash(reader)

		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 64)
	})

	t.Run("with MD5", func(t *testing.T) {
		hasher := util.NewGenericHasher(md5.New)
		data := []byte("test data")
		reader := bytes.NewReader(data)

		hash, err := hasher.CalculateHash(reader)

		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 32) // MD5 produces 32 hex characters
	})

	t.Run("empty data with generic hasher", func(t *testing.T) {
		hasher := util.NewGenericHasher(sha256.New)
		reader := bytes.NewReader([]byte{})

		hash, err := hasher.CalculateHash(reader)

		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
	})

	t.Run("reader error with generic hasher", func(t *testing.T) {
		hasher := util.NewGenericHasher(sha256.New)
		reader := &errorReader{err: errors.New("generic read failed")}

		hash, err := hasher.CalculateHash(reader)

		assert.Error(t, err)
		assert.Empty(t, hash)
	})
}

// errorReader is a mock reader that always returns an error
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}
