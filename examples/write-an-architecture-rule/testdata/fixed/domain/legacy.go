package domain

import "net/http"

// legacyClient predates the layering rule and is allowlisted.
var legacyClient = http.DefaultClient
