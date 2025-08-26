package utils

import "time"

func FormatDate(dateStr string) string {
	if dateStr == "" {
		return "Unknown"
	}
	// Try to parse the date and format it nicely
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t.Format("2006-01-02 15:04")
	}
	return dateStr
}
