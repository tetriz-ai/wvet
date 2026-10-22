// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package domain

import "net/http"

// legacyClient is an allowlisted, pre-existing violation.
var legacyClient = http.DefaultClient
