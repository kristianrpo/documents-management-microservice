package util

import (
	"path/filepath"
	"strings"
)

type MimeTypeDetector interface {
	DetectFromFilename(filename string) string
}

type ExtensionBasedDetector struct {
	extensionMap map[string]string
	defaultType  string
}

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
		defaultType: "application/octet-stream",
	}
}

func (d *ExtensionBasedDetector) DetectFromFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	
	if mimeType, exists := d.extensionMap[ext]; exists {
		return mimeType
	}
	
	return d.defaultType
}

func (d *ExtensionBasedDetector) AddExtension(extension, mimeType string) {
	d.extensionMap[strings.ToLower(extension)] = mimeType
}

type HybridDetector struct {
	detectors []MimeTypeDetector
}

func NewHybridDetector(detectors ...MimeTypeDetector) MimeTypeDetector {
	return &HybridDetector{detectors: detectors}
}

func (h *HybridDetector) DetectFromFilename(filename string) string {
	for _, detector := range h.detectors {
		if mimeType := detector.DetectFromFilename(filename); mimeType != "" && mimeType != "application/octet-stream" {
			return mimeType
		}
	}
	return "application/octet-stream"
}
