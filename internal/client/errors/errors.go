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

type HealthCheckFailedError struct {
	Status string
}

func (e HealthCheckFailedError) Error() string {
	return fmt.Sprintf("Health check не прошел, возвращенный статус: %v", e.Status)
}
