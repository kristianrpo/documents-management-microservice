package util_test

import (
	"testing"

	"github.com/kristianrpo/document-management-microservice/internal/application/util"
	"github.com/stretchr/testify/assert"
)

func TestExtensionBasedDetector_DetectFromFilename(t *testing.T) {
	detector := util.NewExtensionBasedDetector()

	tests := []struct {
		name         string
		filename     string
		expectedMime string
	}{
		{
			name:         "PDF file",
			filename:     "document.pdf",
			expectedMime: "application/pdf",
		},
		{
			name:         "PNG image",
			filename:     "image.png",
			expectedMime: "image/png",
		},
		{
			name:         "JPEG image with .jpg",
			filename:     "photo.jpg",
			expectedMime: "image/jpeg",
		},
		{
			name:         "JPEG image with .jpeg",
			filename:     "photo.jpeg",
			expectedMime: "image/jpeg",
		},
		{
			name:         "GIF image",
			filename:     "animation.gif",
			expectedMime: "image/gif",
		},
		{
			name:         "text file",
			filename:     "readme.txt",
			expectedMime: "text/plain",
		},
		{
			name:         "CSV file",
			filename:     "data.csv",
			expectedMime: "text/csv",
		},
		{
			name:         "JSON file",
			filename:     "config.json",
			expectedMime: "application/json",
		},
		{
			name:         "XML file",
			filename:     "data.xml",
			expectedMime: "application/xml",
		},
		{
			name:         "ZIP file",
			filename:     "archive.zip",
			expectedMime: "application/zip",
		},
		{
			name:         "DOC file",
			filename:     "document.doc",
			expectedMime: "application/msword",
		},
		{
			name:         "DOCX file",
			filename:     "document.docx",
			expectedMime: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		},
		{
			name:         "XLS file",
			filename:     "spreadsheet.xls",
			expectedMime: "application/vnd.ms-excel",
		},
		{
			name:         "XLSX file",
			filename:     "spreadsheet.xlsx",
			expectedMime: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		},
		{
			name:         "unknown extension",
			filename:     "file.xyz",
			expectedMime: "application/octet-stream",
		},
		{
			name:         "no extension",
			filename:     "README",
			expectedMime: "application/octet-stream",
		},
		{
			name:         "uppercase extension",
			filename:     "document.PDF",
			expectedMime: "application/pdf",
		},
		{
			name:         "mixed case extension",
			filename:     "image.PnG",
			expectedMime: "image/png",
		},
		{
			name:         "multiple dots in filename",
			filename:     "my.file.name.pdf",
			expectedMime: "application/pdf",
		},
		{
			name:         "empty filename",
			filename:     "",
			expectedMime: "application/octet-stream",
		},
		{
			name:         "path with extension",
			filename:     "/path/to/file.json",
			expectedMime: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.DetectFromFilename(tt.filename)
			assert.Equal(t, tt.expectedMime, result)
		})
	}
}

func TestExtensionBasedDetector_AddExtension(t *testing.T) {
	detector := util.NewExtensionBasedDetector().(*util.ExtensionBasedDetector)

	// Add a custom extension
	detector.AddExtension(".custom", "application/x-custom")

	result := detector.DetectFromFilename("file.custom")
	assert.Equal(t, "application/x-custom", result)
}

func TestExtensionBasedDetector_AddExtension_OverrideExisting(t *testing.T) {
	detector := util.NewExtensionBasedDetector().(*util.ExtensionBasedDetector)

	// Override an existing extension
	detector.AddExtension(".pdf", "application/x-custom-pdf")

	result := detector.DetectFromFilename("file.pdf")
	assert.Equal(t, "application/x-custom-pdf", result)
}

func TestExtensionBasedDetector_AddExtension_CaseInsensitive(t *testing.T) {
	detector := util.NewExtensionBasedDetector().(*util.ExtensionBasedDetector)

	// Add extension with uppercase
	detector.AddExtension(".CUSTOM", "application/x-custom")

	// Should work with lowercase
	result := detector.DetectFromFilename("file.custom")
	assert.Equal(t, "application/x-custom", result)

	// And with uppercase
	result = detector.DetectFromFilename("file.CUSTOM")
	assert.Equal(t, "application/x-custom", result)
}

func TestHybridDetector_DetectFromFilename(t *testing.T) {
	detector1 := util.NewExtensionBasedDetector()
	detector2 := util.NewExtensionBasedDetector().(*util.ExtensionBasedDetector)
	detector2.AddExtension(".custom", "application/x-custom")

	hybridDetector := util.NewHybridDetector(detector1, detector2)

	tests := []struct {
		name         string
		filename     string
		expectedMime string
	}{
		{
			name:         "found in first detector",
			filename:     "file.pdf",
			expectedMime: "application/pdf",
		},
		{
			name:         "found in second detector",
			filename:     "file.custom",
			expectedMime: "application/x-custom",
		},
		{
			name:         "not found in any detector",
			filename:     "file.unknown",
			expectedMime: "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hybridDetector.DetectFromFilename(tt.filename)
			assert.Equal(t, tt.expectedMime, result)
		})
	}
}

func TestHybridDetector_EmptyDetectorList(t *testing.T) {
	hybridDetector := util.NewHybridDetector()

	result := hybridDetector.DetectFromFilename("file.pdf")
	assert.Equal(t, "application/octet-stream", result)
}
