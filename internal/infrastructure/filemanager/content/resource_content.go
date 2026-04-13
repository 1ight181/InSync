package domain

import "io"

type ResourceContent struct {
	FullPath     string
	RelativePath string

	OpenContent func(fullPath string, relativePath string) (io.ReadCloser, error)
}
