package domain

import (
	"errors"
)

type Path string

var (
	ErrPathEmpty = errors.New("Path не может быть пустым")
)

func NewPath(path string) (Path, error) {
	if path == "" {
		return "", ErrPathEmpty
	}
	return Path(path), nil
}
func (p Path) String() string {
	return string(p)
}
