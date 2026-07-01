package store

import "errors"

// ErrNotFound is returned when a short URL has no mapping in either the cache or the durable store.
var ErrNotFound = errors.New("short url not found")
