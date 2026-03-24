package resolver

import (
	"fmt"
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
