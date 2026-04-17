package hash

type IHashCache interface {
	LoadHashCache(hashSet map[string]string) error
	GetHashCache(fullPath string) (string, error)
	SetHashCache(fullPath string, hash string)
}
