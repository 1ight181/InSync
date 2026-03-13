package errors

import "fmt"

var (
	ErrServerAlreadyStarted = fmt.Errorf("mDNS сервер уже запущен")
	ErrServerAlreadyStopped = fmt.Errorf("mDNS сервер уже остановлен")
)

type FailedToGetAddressesError struct {
	Err       error
	Interface string
}

func (e FailedToGetAddressesError) Error() string {
	return fmt.Sprintf("Не удалось получить адреса для интерфейса %s: %v", e.Interface, e.Err)
}

func (e FailedToGetAddressesError) Unwrap() error {
	return e.Err
}

type FailedToGetNetworkInterfaceByNameError struct {
	Err       error
	Interface string
}

func (e FailedToGetNetworkInterfaceByNameError) Error() string {
	return fmt.Sprintf("Не удалось получить интерфейс по имени %s: %v", e.Interface, e.Err)
}

func (e FailedToGetNetworkInterfaceByNameError) Unwrap() error {
	return e.Err
}

type FailedToStartMDnsServerError struct {
	Err error
}

func (e FailedToStartMDnsServerError) Error() string {
	return fmt.Sprintf("Не удалось запустить mDNS: %v", e.Err)
}

func (e FailedToStartMDnsServerError) Unwrap() error {
	return e.Err
}
