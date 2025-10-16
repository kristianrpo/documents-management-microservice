package util_test

import (
	"testing"

	"github.com/kristianrpo/document-management-microservice/internal/application/util"
	"github.com/stretchr/testify/assert"
)

func TestObjectKeyFromHash(t *testing.T) {
	tests := []struct {
		name        string
		hashHex     string
		filename    string
		expectedKey string
	}{
		{
			name:        "valid hash and filename with extension",
			hashHex:     "a3b2c1d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
			filename:    "document.pdf",
			expectedKey: "a3/a3b2c1d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2.pdf",
		},
		{
			name:        "filename with multiple dots",
			hashHex:     "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			filename:    "my.file.name.jpg",
			expectedKey: "12/1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef.jpg",
		},
		{
			name:        "filename without extension",
			hashHex:     "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			filename:    "README",
			expectedKey: "ab/abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		},
		{
			name:        "uppercase extension gets lowercased",
			hashHex:     "fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
			filename:    "document.PDF",
			expectedKey: "fe/fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321.pdf",
		},
		{
			name:        "mixed case extension gets lowercased",
			hashHex:     "112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00",
			filename:    "image.PnG",
			expectedKey: "11/112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00.png",
		},
		{
			name:        "short hash less than 2 characters",
			hashHex:     "a",
			filename:    "file.txt",
			expectedKey: "00/a.txt",
		},
		{
			name:        "empty hash",
			hashHex:     "",
			filename:    "file.txt",
			expectedKey: "00/.txt",
		},
		{
			name:        "hash with exactly 2 characters",
			hashHex:     "ab",
			filename:    "test.json",
			expectedKey: "ab/ab.json",
		},
		{
			name:        "filename with leading dot",
			hashHex:     "aabbccdd00112233aabbccdd00112233aabbccdd00112233aabbccdd00112233",
			filename:    ".gitignore",
			expectedKey: "aa/aabbccdd00112233aabbccdd00112233aabbccdd00112233aabbccdd00112233.gitignore",
		},
		{
			name:        "filename ending with dot",
			hashHex:     "112233aabbccdd00112233aabbccdd00112233aabbccdd00112233aabbccdd00",
			filename:    "file.",
			expectedKey: "11/112233aabbccdd00112233aabbccdd00112233aabbccdd00112233aabbccdd00.",
		},
		{
			name:        "special characters in filename",
			hashHex:     "aabbccddee112233aabbccddee112233aabbccddee112233aabbccddee112233",
			filename:    "my-file_name (1).pdf",
			expectedKey: "aa/aabbccddee112233aabbccddee112233aabbccddee112233aabbccddee112233.pdf",
		},
		{
			name:        "common image extension",
			hashHex:     "ffeeddccbbaa9988ffeeddccbbaa9988ffeeddccbbaa9988ffeeddccbbaa9988",
			filename:    "photo.jpeg",
			expectedKey: "ff/ffeeddccbbaa9988ffeeddccbbaa9988ffeeddccbbaa9988ffeeddccbbaa9988.jpeg",
		},
		{
			name:        "compressed file extension",
			hashHex:     "1122334455667788990011223344556677889900112233445566778899001122",
			filename:    "archive.zip",
			expectedKey: "11/1122334455667788990011223344556677889900112233445566778899001122.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.ObjectKeyFromHash(tt.hashHex, tt.filename)
			assert.Equal(t, tt.expectedKey, result)
		})
	}
}
