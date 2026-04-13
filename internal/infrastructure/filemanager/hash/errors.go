package hash

import "fmt"

type HashCacheNotFoundError struct {
	FullPath string
}

func (e HashCacheNotFoundError) Error() string {
	return fmt.Sprintf("Кэш хэша не найдена для: %s", e.FullPath)
}
