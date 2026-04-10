package resolver

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidScheme         = errors.New("while parse target: invalid scheme")
	ErrMissingServiceName    = errors.New("while parse target: missing service name")
	ErrMissingEndpoint       = errors.New("while parse target: missing endpoint")
	ErrInvalidEndpointFormat = errors.New("while parse target: invalid endpoint format")
)

type FailedToGetNetworkInterfaceByNameError struct {
	Err           error
	InterfaceName string
}

func (e FailedToGetNetworkInterfaceByNameError) Error() string {
	return fmt.Sprintf("Failed to get network interface by name %s: %v", e.InterfaceName, e.Err)
}

func (e FailedToGetNetworkInterfaceByNameError) Unwrap() error {
	return e.Err
}
