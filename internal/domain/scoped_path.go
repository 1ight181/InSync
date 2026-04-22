package domain

import (
	"errors"
	"path/filepath"
	"strings"
)

type ScopedPath struct {
	Root RootName
	Path Path
}

var (
	ErrAbsPath = errors.New("путь не может быть абсолютным")
)

func NewScopedPath(root RootName, path Path) (ScopedPath, error) {
	if filepath.IsAbs(path.String()) || strings.HasPrefix(path.String(), "/") {
		return ScopedPath{}, ErrAbsPath
	}

	return ScopedPath{Root: root, Path: path}, nil
}
