// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package domain

import "net/http"

// Order knows about HTTP, which belongs in the transport layer.
type Order struct{ Status int }

func (o Order) OK() bool { return o.Status == http.StatusOK }
