package errors

import "errors"

// Ошибки лоудера конфигурации
type FailedToLoadConfigError struct {
	Err error
}

func (e *FailedToLoadConfigError) Error() string {
	return "Не удалось загрузить конфигурацию: " + e.Err.Error()
}

// Ошибки валидации конфигурации

// Ошибки валидации конфигурации сервера
var (
	ErrCertDirIsEmpty       = errors.New("Директория сертификата не может быть пустой")
	ErrCertFilenameIsEmpty  = errors.New("Имя файла сертификата не может быть пустым")
	ErrCertExtensionIsEmpty = errors.New("Расширение файла сертификата не может быть пустым")

	ErrServerIpIsEmpty          = errors.New("IP сервера не может быть пустым")
	ErrServerPortIsInvalid      = errors.New("Порт сервера должен быть целым числом в диапозоне от 0 до 65535")
	ErrServerNetworkTypeIsEmpty = errors.New("Тип сети сервера не может быть пустым")
)

type CertFileDoesNotExistError struct {
	FilePath string
}

func (e *CertFileDoesNotExistError) Error() string {
	return "Файл сертификата не существует: " + e.FilePath
}

// Ошибки валидации конфигурации базы данных
var (
	ErrDbHostIsEmpty     = errors.New("Хост БД не может быть пустым")
	ErrDbPortIsInvalid   = errors.New("Порт БД должен быть целым числом в диапозоне от 0 до 65535")
	ErrDbUserIsEmpty     = errors.New("Пользователь БД не может быть пустым")
	ErrDbPasswordIsEmpty = errors.New("Пароль БД не может быть пустым")
)
