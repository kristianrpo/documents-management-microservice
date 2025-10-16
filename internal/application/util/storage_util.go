package util

import (
	"fmt"
	"strings"
)

// ObjectKeyFromHash generates an S3 object key from a file hash and filename
// The key includes a prefix from the first 2 characters of the hash for better S3 performance
func ObjectKeyFromHash(hashHex, filename string) string {
	ext := ""
	if dot := strings.LastIndex(filename, "."); dot >= 0 {
		ext = strings.ToLower(filename[dot:])
	}
	prefix := "00"
	if len(hashHex) >= 2 {
		prefix = hashHex[:2]
	}
	return fmt.Sprintf("%s/%s%s", prefix, hashHex, ext)
}
