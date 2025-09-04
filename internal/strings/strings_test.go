package strings

import "testing"

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		expected  string
	}{
		{
			name:      "short",
			input:     "hello",
			maxLength: 10,
			expected:  "hello",
		},
		{
			name:      "equal",
			input:     "hello world",
			maxLength: 11,
			expected:  "hello world",
		},
		{
			name:      "long",
			input:     "this is a long string",
			maxLength: 10,
			expected:  "this is...",
		},
		{
			name:      "empty",
			input:     "",
			maxLength: 5,
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateString(tt.input, tt.maxLength)
			if result != tt.expected {
				t.Errorf("TruncateString(%q, %d) = %q, expected %q", tt.input, tt.maxLength, result, tt.expected)
			}
		})
	}
}
