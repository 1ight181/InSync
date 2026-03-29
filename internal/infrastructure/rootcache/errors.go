package rootcache

import (
	"fmt"
)

type RootCacheNotFoundError struct {
	RootName string
}

func (e RootCacheNotFoundError) Error() string {
	return fmt.Sprintf("root не найден для: %s", e.RootName)
}
