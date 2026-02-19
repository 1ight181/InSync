package errors

import (
	"errors"
)

var (
	ErrFailedToAppendCa     = errors.New("не удалось добавить CA сертификат в пул")
	ErrClientAlreadyStarted = errors.New("клиент уже запущен")
	ErrClientAlreadyStopped = errors.New("клиент уже остановлен")
	ErrClientNotStarted     = errors.New("клиент не запущен")
)

type DeleteFileFailedError struct {
	Message string
}

func (e DeleteFileFailedError) Error() string {
	return e.Message
}

type RenameFileFailedError struct {
	Message string
}

func (e RenameFileFailedError) Error() string {
	return e.Message
}

type PutFileFailedError struct {
	Message string
}

func (e PutFileFailedError) Error() string {
	return e.Message
}
