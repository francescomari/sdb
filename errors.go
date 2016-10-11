package sdb

import "errors"

// ErrInvalidFormat is the error returned from various function if the requested
// output format is not supported or invalid.
var ErrInvalidFormat = errors.New("Invalid format")
