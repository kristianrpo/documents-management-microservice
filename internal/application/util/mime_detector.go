package util

import (
	"path/filepath"
	"strings"
)

const (
	// DefaultMimeType is the fallback MIME type for unknown file types
	DefaultMimeType = "application/octet-stream"
)

// MimeTypeDetector defines the interface for detecting MIME types from filenames
type MimeTypeDetector interface {
	DetectFromFilename(filename string) string
}

// ExtensionBasedDetector implements MimeTypeDetector by mapping file extensions to MIME types
type ExtensionBasedDetector struct {
	extensionMap map[string]string
	defaultType  string
}

// NewExtensionBasedDetector creates a new extension-based MIME type detector with common file types
func NewExtensionBasedDetector() MimeTypeDetector {
	return &ExtensionBasedDetector{
		extensionMap: map[string]string{
			".pdf":  "application/pdf",
			".png":  "image/png",
			".jpg":  "image/jpeg",
			".jpeg": "image/jpeg",
			".gif":  "image/gif",
			".txt":  "text/plain",
			".csv":  "text/csv",
			".json": "application/json",
			".xml":  "application/xml",
			".zip":  "application/zip",
			".doc":  "application/msword",
			".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			".xls":  "application/vnd.ms-excel",
			".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		},
		defaultType: DefaultMimeType,
	}
}

// DetectFromFilename detects the MIME type based on the file extension
func (d *ExtensionBasedDetector) DetectFromFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	if mimeType, exists := d.extensionMap[ext]; exists {
		return mimeType
	}

	return d.defaultType
}

// AddExtension adds a custom file extension to MIME type mapping
func (d *ExtensionBasedDetector) AddExtension(extension, mimeType string) {
	d.extensionMap[strings.ToLower(extension)] = mimeType
}

// HybridDetector combines multiple MIME type detectors, trying each in sequence
type HybridDetector struct {
	detectors []MimeTypeDetector
}

// NewHybridDetector creates a new hybrid detector that tries multiple detection strategies
func NewHybridDetector(detectors ...MimeTypeDetector) MimeTypeDetector {
	return &HybridDetector{detectors: detectors}
}

// DetectFromFilename tries each detector in sequence until a specific MIME type is found
func (h *HybridDetector) DetectFromFilename(filename string) string {
	for _, detector := range h.detectors {
		if mimeType := detector.DetectFromFilename(filename); mimeType != "" && mimeType != DefaultMimeType {
			return mimeType
		}
	}
	return DefaultMimeType
}
