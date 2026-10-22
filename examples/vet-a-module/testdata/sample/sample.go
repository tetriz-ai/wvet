// Package sample has one of each problem wvet reports.
package sample

import (
	"errors"
	"sync"
)

func load() (string, error) { return "", errors.New("unavailable") }

// Value checks the error, then drops it for a new one.
func Value() (string, error) {
	v, err := load()
	if err != nil {
		return "", errors.New("no value")
	}
	return v, nil
}

// StartAll starts each job in a goroutine and waits for all of them.
func StartAll(jobs []func()) {
	var wg sync.WaitGroup
	for _, job := range jobs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			job()
		}()
	}
	wg.Wait()
}
