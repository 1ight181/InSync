package domain

import "io"

type ResourceContent struct {
	Path        string
	OpenContent func(fullPath string) (io.ReadCloser, error)
}
