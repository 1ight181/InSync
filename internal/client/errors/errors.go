package errors

import (
	"errors"
	"fmt"
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
	return fmt.Sprintf("DeleteFile не удался: %v", e.Message)
}

type RenameFileFailedError struct {
	Message string
}

func (e RenameFileFailedError) Error() string {
	return fmt.Sprintf("RenameFile не удался: %v", e.Message)
}

type PutFileFailedError struct {
	Message string
}

func (e PutFileFailedError) Error() string {
	return fmt.Sprintf("PutFile не удался: %v", e.Message)
}

type FailedToStartHealthCheckerError struct {
	Err error
}

func (e FailedToStartHealthCheckerError) Error() string {
	return fmt.Sprintf("Не удалось запустить HealthChecker: %v", e.Err)
}
