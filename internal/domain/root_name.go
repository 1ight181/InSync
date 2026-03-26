package domain

import "errors"

type RootName string

var (
	ErrRootNameEmpty = errors.New("RootName не может быть пустым")
)

func NewRootName(name string) (RootName, error) {
	if name == "" {
		return "", ErrRootNameEmpty
	}
	return RootName(name), nil
}
