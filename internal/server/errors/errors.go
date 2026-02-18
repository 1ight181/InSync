package errors

import (
	"errors"
	"fmt"
)

type ServerStartError struct {
	Err error
}

func (e ServerStartError) Error() string {
	return fmt.Sprintf("Не удалось запустить сервер: %v", e.Err)
}

var ErrFailedToAppendCa = errors.New("не удалось добавить CA сертификат в пул")
