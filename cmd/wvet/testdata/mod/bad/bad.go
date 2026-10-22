// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

// Package bad breaks each of the three rules once.
package bad

import (
	"errors"
	"sync"
)

func load() (string, error) { return "", errors.New("unavailable") }

// Value checks the error and drops it.
func Value() (string, error) {
	v, err := load()
	if err != nil {
		return "", errors.New("no value")
	}
	return v, nil
}

// Start starts a goroutine that does not recover.
func Start(work func()) {
	go work()
}

// Drain waits with no bound.
func Drain(wg *sync.WaitGroup) {
	wg.Wait()
}
