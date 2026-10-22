// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package domain

// Order holds no transport types.
type Order struct{ Paid bool }

func (o Order) OK() bool { return o.Paid }
