package hash

type HashCacheEntry struct {
	Id       string `gorm:"primaryKey;type:uuid"`
	Hash     string `gorm:"type:varchar(255);uniqueIndex:ux_cache_hash_path"`
	FullPath string
}
