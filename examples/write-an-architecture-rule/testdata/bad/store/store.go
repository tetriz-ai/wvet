package store

import "fmt"

// Save prints instead of returning.
func Save(key string) { fmt.Println("saved", key) }
