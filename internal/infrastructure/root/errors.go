package root

import (
	"fmt"
)

type RootNotFoundError struct {
	RootName string
}

func (e RootNotFoundError) Error() string {
	return fmt.Sprintf("root не найден для: %s", e.RootName)
}
