package hash

import "fmt"

type FileInfoCacheNotFoundError struct {
	FullPath string
}

func (e FileInfoCacheNotFoundError) Error() string {
	return fmt.Sprintf("Информация о файле не найдена в кэше для: %s", e.FullPath)
}
