package parse

// MustInt has the one panic the ratchet allows.
func MustInt(s string) int {
	if s == "" {
		panic("empty")
	}
	return len(s)
}
