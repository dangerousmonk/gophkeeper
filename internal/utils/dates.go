package utils

import "time"

// FormatDate tries to parse date string to RFC3339 format.In case of error return original string or unknown If provided string is empty
func FormatDate(dateStr string) string {
	if dateStr == "" {
		return "Unknown"
	}
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t.Format("2006-01-02 15:04:05 MST")
	}
	return dateStr
}
