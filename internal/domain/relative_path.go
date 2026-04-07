package domain

import "errors"

type RelativePath string

var (
	ErrRelativePathEmpty = errors.New("RelativePath не может быть пустым")
)

func NewRelativePath(path string) (RelativePath, error) {
	if path == "" {
		return "", ErrRelativePathEmpty
	}
	return RelativePath(path), nil
}

func (rp RelativePath) String() string {
	return string(rp)
}
