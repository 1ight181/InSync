package hash

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type HashCacheEntry struct {
	Id       string `gorm:"primaryKey;type:uuid"`
	Hash     string `gorm:"type:varchar(255);uniqueIndex:ux_cache_hash_path"`
	FullPath string
}

func (e *HashCacheEntry) BeforeCreate(tx *gorm.DB) error {
	e.Id = uuid.NewString()
	return nil
}
