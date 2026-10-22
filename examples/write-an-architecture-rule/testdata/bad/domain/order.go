package domain

import "net/http"

// Order knows an HTTP status, which belongs in the transport layer.
type Order struct{ Status int }

func (o Order) OK() bool { return o.Status == http.StatusOK }
