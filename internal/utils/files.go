package utils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ByteSize uint64

const (
	B  ByteSize = 1
	KB          = B << 10
	MB          = KB << 10
	GB          = MB << 10
	TB          = GB << 10
)

// FormatFileSize is a helper to format file size from bytes to other units
func FormatFileSize(size int64) string {
	switch {
	case size < int64(KB):
		return fmt.Sprintf("%d bytes", size)
	case size < int64(MB):
		return fmt.Sprintf("%.1f KB", float64(size)/float64(KB))
	case size < int64(GB):
		return fmt.Sprintf("%.1f MB", float64(size)/float64(MB))
	case size < int64(TB):
		return fmt.Sprintf("%.1f GB", float64(size)/float64(GB))
	default:
		return fmt.Sprintf("%d bytes", size)
	}
}

// GetFileMetadata returns file metadata including its name, path, size and it's type as a string
func GetFileMetadata(path string) (map[string]any, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	fileType := strings.TrimPrefix(filepath.Ext(path), ".")
	if fileType == "" {
		fileType = "unknown"
	}
	meta := map[string]any{
		"file_name": fileInfo.Name(),
		"file_path": path,
		"file_size": float64(fileInfo.Size()),
		"file_type": fileType,
	}
	return meta, nil
}

// MergeChunks combines chunks into the original byte array
func MergeChunks(chunks [][]byte) []byte {
	var buffer bytes.Buffer
	for _, chunk := range chunks {
		if chunk != nil {
			buffer.Write(chunk)
		}
	}

	return buffer.Bytes()
}
