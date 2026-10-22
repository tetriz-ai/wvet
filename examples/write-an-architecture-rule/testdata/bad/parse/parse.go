package parse

// MustInt has a second panic, one more than the ratchet allows.
func MustInt(s string) int {
	if s == "" {
		panic("empty")
	}
	if s[0] == '-' {
		panic("negative")
	}
	return len(s)
}
