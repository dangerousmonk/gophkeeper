package utils

import (
	"testing"
	"time"
)

func TestFormatDate(t *testing.T) {
	now := time.Now()
	currentRFC3339 := now.Format(time.RFC3339)
	currentFormatted := now.Format("2006-01-02 15:04:05 MST")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty_string",
			input:    "",
			expected: "Unknown",
		},
		{
			name:     "valid_RFC3339_UTC",
			input:    "2023-12-25T15:30:45Z",
			expected: "2023-12-25 15:30:45 UTC",
		},
		{
			name:     "RFC3339_positive_tz",
			input:    "2023-12-25T15:30:45+03:00",
			expected: "2023-12-25 15:30:45 MSK",
		},
		{
			name:     "RFC3339_with_milliseconds",
			input:    "2023-12-25T15:30:45.123Z",
			expected: "2023-12-25 15:30:45 UTC",
		},
		{
			name:     "current_time_RFC3339",
			input:    currentRFC3339,
			expected: currentFormatted,
		},
		{
			name:     "invalid_format",
			input:    "2023/12/25 15:30:45",
			expected: "2023/12/25 15:30:45",
		},
		{
			name:     "malformed_date_string",
			input:    "not-a-date",
			expected: "not-a-date",
		},
		{
			name:     "date_without_time",
			input:    "2023-12-25",
			expected: "2023-12-25",
		},
		{
			name:     "ISO8601_format",
			input:    "2023-12-25T15:30:45+0300",
			expected: "2023-12-25T15:30:45+0300",
		},
		{
			name:     "whitespace_only",
			input:    "   ",
			expected: "   ",
		},
		{
			name:     "RFC3339_midnight",
			input:    "2023-12-25T00:00:00Z",
			expected: "2023-12-25 00:00:00 UTC",
		},
		{
			name:     "RFC3339_day_end",
			input:    "2023-12-25T23:59:59Z",
			expected: "2023-12-25 23:59:59 UTC",
		},
		{
			name:     "RFC3339_with_EST_tz",
			input:    "2023-12-25T15:30:45-05:00",
			expected: "2023-12-25 15:30:45 -0500",
		},
		{
			name:     "RFC3339_with_CET_tz",
			input:    "2023-12-25T15:30:45+01:00",
			expected: "2023-12-25 15:30:45 +0100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDate(tt.input)
			if result != tt.expected {
				t.Errorf("FormatDate(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
