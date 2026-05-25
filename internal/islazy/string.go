// string.go: Text utility functions such as string truncation.
package islazy

// LeftTrucate a string if its more than max
func LeftTrucate(s string, max int) string {
	if len(s) <= max {
		return s
	}

	return s[max:]
}

// Truncate shortens a string when it exceeds maxLength (by bytes).
func Truncate(s string, maxLength int) string {
	if maxLength <= 0 || len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}
