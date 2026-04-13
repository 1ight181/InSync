package root

import (
	"fmt"
	"insync/internal/domain"
)

type RootNotFoundError struct {
	RootName domain.RootName
}

func (e RootNotFoundError) Error() string {
	return fmt.Sprintf("root не найден для: %s", e.RootName)
}
