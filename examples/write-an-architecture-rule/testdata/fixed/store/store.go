package store

import "fmt"

// Save reports what it did to its caller.
func Save(key string) string { return fmt.Sprintf("saved %s", key) }
