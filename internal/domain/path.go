package domain

import (
	"errors"
	"path/filepath"
	"strings"
)

type Path string

var (
	ErrPathEmpty            = errors.New("Path не может быть пустым")
	ErrCantJoinDifferentAbs = errors.New("Нельзя объединить разные абсолютные пути")
)

func NewPath(path string) (Path, error) {
	if path == "" {
		return "", ErrPathEmpty
	}

	path = filepath.Clean(path)

	return Path(path), nil
}
func (p Path) String() string {
	return string(p)
}

func (p Path) IsEmpty() bool {
	return p == ""
}

func (p Path) Abs() (Path, error) {
	abs, err := filepath.Abs(p.String())
	if err != nil {
		return "", err
	}

	return Path(abs), nil
}

func (p Path) Dir() Path {
	return Path(filepath.Dir(p.String()))
}

func (p Path) Join(newPath string) (Path, error) {
	newPath = filepath.Clean(newPath)

	if filepath.IsAbs(newPath) {
		path, found := strings.CutPrefix(p.String(), newPath)
		if found {
			return Path(filepath.Join(newPath, path)), nil
		}

		return "", ErrCantJoinDifferentAbs
	}

	return Path(filepath.Join(p.String(), newPath)), nil
}
