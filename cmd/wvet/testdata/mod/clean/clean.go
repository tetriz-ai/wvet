// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

// Package clean breaks none of the rules.
package clean

import (
	"errors"
	"fmt"
)

func load() (string, error) { return "", errors.New("unavailable") }

// Value wraps and returns the error.
func Value() (string, error) {
	v, err := load()
	if err != nil {
		return "", fmt.Errorf("loading the value: %w", err)
	}
	return v, nil
}
