package resolver

import (
	"errors"
)

var (
	ErrUnsupportedScheme = errors.New("unsupported scheme: expected mdns")
	ErrEmptyEndpoint     = errors.New("empty endpoint")
	ErrInvalidFormat     = errors.New("invalid mdns format")
	ErrInvalidService    = errors.New("invalid service format")
)
