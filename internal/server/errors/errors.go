package errors

import "fmt"

type ServerStartError struct {
	Err error
}

func (e ServerStartError) Error() string {
	return fmt.Sprintf("Не удалось запустить сервер: %v", e.Err)
}

var ErrFailedToAppendCa = fmt.Errorf("не удалось добавить CA сертификат в пул")
