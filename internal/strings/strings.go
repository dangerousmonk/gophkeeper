package strings

// TruncateString truncates string to the maxLength with ellipsis at the end. Returns original string If len of s less or equal maxLength
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength-3] + "..."
}
