package pathtree

import (
	"errors"
)

var (
	ErrNotFound       = errors.New("Путь не найден в дереве")
	ErrAbsPath        = errors.New("Пусть не может быть абсолютным")
	ErrParentNotFound = errors.New("Родительского узла не существует")
)
