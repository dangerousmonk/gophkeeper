package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMergeChunks(t *testing.T) {
	tests := []struct {
		name     string
		chunks   [][]byte
		expected []byte
	}{
		{
			name:     "nil",
			chunks:   nil,
			expected: []byte{},
		},
		{
			name:     "empty",
			chunks:   [][]byte{},
			expected: []byte{},
		},
		{
			name:     "single_nil",
			chunks:   [][]byte{nil},
			expected: []byte{},
		},
		{
			name:     "multiple_nil",
			chunks:   [][]byte{nil, nil, nil},
			expected: []byte{},
		},
		{
			name:     "single_empty_chunk",
			chunks:   [][]byte{{}},
			expected: []byte{},
		},
		{
			name:     "multiple_empty_chunks",
			chunks:   [][]byte{{}, {}, {}},
			expected: []byte{},
		},
		{
			name:     "single_non_empty_chunk",
			chunks:   [][]byte{{1, 2, 3}},
			expected: []byte{1, 2, 3},
		},
		{
			name:     "multiple_non_empty_chunks",
			chunks:   [][]byte{{1, 2}, {3, 4}, {5, 6}},
			expected: []byte{1, 2, 3, 4, 5, 6},
		},
		{
			name:     "mixed_nil_non_empty_chunks",
			chunks:   [][]byte{nil, {1, 2}, nil, {3, 4}, nil},
			expected: []byte{1, 2, 3, 4},
		},
		{
			name:     "mixed_empty_non_empty_chunks",
			chunks:   [][]byte{{}, {1, 2}, {}, {3, 4}, {}},
			expected: []byte{1, 2, 3, 4},
		},
		{
			name:     "large_chunks",
			chunks:   [][]byte{make([]byte, 1024), make([]byte, 2048)},
			expected: make([]byte, 3072), // 1024 + 2048
		},
		{
			name:     "chunks_with_zero_bytes",
			chunks:   [][]byte{{0, 0, 0}, {0, 0}},
			expected: []byte{0, 0, 0, 0, 0},
		},
		{
			name:     "chunks_with_maximum_byte_values",
			chunks:   [][]byte{{255, 255}, {255}},
			expected: []byte{255, 255, 255},
		},
		{
			name:     "chunks_with_mixed_byte_values",
			chunks:   [][]byte{{'a', 'b'}, {'c'}, {'d', 'e', 'f'}},
			expected: []byte{'a', 'b', 'c', 'd', 'e', 'f'},
		},
		{
			name:     "nil_chunks_in_middle",
			chunks:   [][]byte{{1, 2}, nil, {3, 4}, nil, {5, 6}},
			expected: []byte{1, 2, 3, 4, 5, 6},
		},
		{
			name:     "all_nil_chunks_with_empty_slices",
			chunks:   [][]byte{nil, {}, nil, {}},
			expected: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeChunks(tt.chunks)
			if len(result) != len(tt.expected) {
				t.Errorf("MergeChunks() length = %d, expected %d", len(result), len(tt.expected))
				return
			}

			for i := range tt.expected {
				if result[i] != tt.expected[i] {
					t.Errorf("MergeChunks() at index %d = %d, expected %d", i, result[i], tt.expected[i])
					break
				}
			}
		})
	}
}

func TestGetFileMetadataOK(t *testing.T) {
	tmpDir := t.TempDir()
	testFiles := setupTestFiles(t, tmpDir)

	tests := []struct {
		name        string
		path        string
		want        map[string]any
		expectError bool
		errContains string
	}{
		{
			name: "regular_text_file",
			path: testFiles["text"],
			want: map[string]any{
				"file_name": "test.txt",
				"file_path": testFiles["text"],
				"file_size": 11.0, // "hello world" = 11 bytes
				"file_type": "txt",
			},
			expectError: false,
		},
		{
			name: "file_with_multiple_extensions",
			path: testFiles["multiExt"],
			want: map[string]any{
				"file_name": "archive.tar.gz",
				"file_path": testFiles["multiExt"],
				"file_size": 0.0,
				"file_type": "gz", // Should get the last extension
			},
			expectError: false,
		},
		{
			name: "file_with_no_extension",
			path: testFiles["noExt"],
			want: map[string]any{
				"file_name": "no_extension",
				"file_path": testFiles["noExt"],
				"file_size": 0.0,
				"file_type": "unknown",
			},
			expectError: false,
		},
		{
			name: "file_with_dot_no_extension",
			path: testFiles["dotNoExt"],
			want: map[string]any{
				"file_name": "file.version",
				"file_path": testFiles["dotNoExt"],
				"file_size": 0.0,
				"file_type": "version",
			},
			expectError: false,
		},
		{
			name: "empty_file",
			path: testFiles["empty"],
			want: map[string]any{
				"file_name": "empty.txt",
				"file_path": testFiles["empty"],
				"file_size": 0.0,
				"file_type": "txt",
			},
			expectError: false,
		},
		{
			name:        "non_existent_file",
			path:        filepath.Join(tmpDir, "non_existent_file.txt"),
			want:        nil,
			expectError: true,
			errContains: "failed to get file info",
		},
		{
			name:        "empty_string_path",
			path:        "",
			want:        nil,
			expectError: true,
			errContains: "failed to get file info",
		},
		{
			name: "file_with_complex_extension",
			path: testFiles["complexExt"],
			want: map[string]any{
				"file_name": "config.env.json",
				"file_path": testFiles["complexExt"],
				"file_size": 0.0,
				"file_type": "json",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFileMetadata(tt.path)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error but got none")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error message %q should contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetFileMetadata() mismatch (-want +got):\n%s", diff)
			}

			if fileName, ok := got["file_name"].(string); !ok || fileName == "" {
				t.Errorf("file_name should be non-empty string, got %v", got["file_name"])
			}
			if filePath, ok := got["file_path"].(string); !ok || filePath != tt.path {
				t.Errorf("file_path should match input path, got %v, want %v", filePath, tt.path)
			}
			if fileSize, ok := got["file_size"].(float64); !ok {
				t.Errorf("file_size should be float64, got %T", got["file_size"])
			} else if fileSize < 0 {
				t.Errorf("file_size should be non-negative, got %f", fileSize)
			}
			if fileType, ok := got["file_type"].(string); !ok {
				t.Errorf("file_type should be string, got %T", fileType)
			}
		})
	}
}

// setupTestFiles creates various test files in the temporary directory
func setupTestFiles(t *testing.T, tmpDir string) map[string]string {
	t.Helper()

	files := map[string]string{
		"text":     filepath.Join(tmpDir, "test.txt"),
		"multiExt": filepath.Join(tmpDir, "archive.tar.gz"),
		"noExt":    filepath.Join(tmpDir, "no_extension"),
		"dotNoExt": filepath.Join(tmpDir, "file.version"),
		"empty":    filepath.Join(tmpDir, "empty.txt"),
		// "onlyDot":    filepath.Join(tmpDir, "."),
		"complexExt": filepath.Join(tmpDir, "config.env.json"),
	}

	if err := os.WriteFile(files["text"], []byte("hello world"), 0644); err != nil {
		t.Fatalf("Failed to create text file: %v", err)
	}

	for k, path := range files {
		if k == "text" {
			continue
		}
		if err := os.WriteFile(path, []byte{}, 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}

	return files
}

func TestGetFileMetadataErrors(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		errContains string
	}{
		{
			name:        "non-existent file",
			path:        "/non/existent/path",
			errContains: "failed to get file info",
		},
		{
			name:        "permission denied directory",
			path:        "/root/protected", // Assuming this would be protected
			errContains: "failed to get file info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetFileMetadata(tt.path)
			if err == nil {
				t.Fatal("Expected error but got none")
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("Error message %q should contain %q", err.Error(), tt.errContains)
			}
		})
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		// Bytes cases
		{
			name:     "zero_bytes",
			size:     0,
			expected: "0 bytes",
		},
		{
			name:     "1_bytes",
			size:     1,
			expected: "1 bytes",
		},
		{
			name:     "1023_bytes",
			size:     1023,
			expected: "1023 bytes",
		},

		// KB cases
		{
			name:     "1_KB",
			size:     1024,
			expected: "1.0 KB",
		},
		{
			name:     "1.5_KB",
			size:     1536, // 1024 + 512
			expected: "1.5 KB",
		},
		{
			name:     "1023.9_KB",
			size:     1024*1024 - 1,
			expected: "1024.0 KB",
		},
		{
			name:     "999.9_KB",
			size:     1024*1000 - 1,
			expected: "1000.0 KB",
		},
		{
			name:     "512.5_KB",
			size:     512*1024 + 512,
			expected: "512.5 KB",
		},

		// MB cases
		{
			name:     "1_MB",
			size:     1024 * 1024,
			expected: "1.0 MB",
		},
		{
			name:     "1.5_MB",
			size:     1024 * 1024 * 3 / 2, // 1.5 MB
			expected: "1.5 MB",
		},
		{
			name:     "2.25_MB",
			size:     1024 * 1024 * 9 / 4, // 2.25 MB
			expected: "2.2 MB",            // 2.25 rounds to 2.2 due to .1f formatting
		},
		{
			name:     "999.9_MB",
			size:     1024*1024*1000 - 1,
			expected: "1000.0 MB",
		},
		{
			name:     "1024_MB",
			size:     1024 * 1024 * 1024,
			expected: "1.0 GB",
		},

		// Edge cases and large numbers
		{
			name:     "negative_size",
			size:     -1,
			expected: "-1 bytes",
		},
		{
			name:     "very_large_size",
			size:     5 * 1024 * 1024 * 1024,
			expected: "5.0 GB",
		},
		// {
		// 	name:     "max_int64_size",
		// 	size:     1<<63 - 1,
		// 	expected: "8388608.0 MB", // 8,388,608 MB = 8,192 GB
		// },

		// Precision testing
		{
			name:     "1.05_KB",
			size:     1075, // 1024 + 51 = 1.05 KB
			expected: "1.0 KB",
		},
		{
			name:     "1.04_KB",
			size:     1065, // 1024 + 41 = 1.04 KB
			expected: "1.0 KB",
		},
		{
			name:     "1.55_MB",
			size:     1625292, // 1.55 MB
			expected: "1.5 MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatFileSize(tt.size)
			if result != tt.expected {
				t.Errorf("FormatFileSize(%d) = %q, expected %q", tt.size, result, tt.expected)
			}
		})
	}
}
