// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package main

import (
	"strings"
	"testing"
)

func TestFixRequested(t *testing.T) {
	tests := []struct {
		args string
		want bool
	}{
		{"-fix ./...", true},
		{"--fix ./...", true},
		{"---fix ./...", true},
		{"-fix=true ./...", true},
		{"-fix=t ./...", true},
		{"-fix=True ./...", true},
		{"./... -fix", true},
		{"-tags -- -fix ./...", true},
		{"-tags -fix ./...", true},
		{"-fix=false ./...", false},
		{"-fix=0 ./...", false},
		{"-fix=yes ./...", false},
		{"-fix= ./...", false},
		{"fix ./...", false},
		{"fix=true ./...", false},
		{"-fixes ./...", false},
		{"./...", false},
	}
	for _, tt := range tests {
		t.Run(tt.args, func(t *testing.T) {
			if got := fixRequested(strings.Fields(tt.args)); got != tt.want {
				t.Errorf("fixRequested(%q) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}
