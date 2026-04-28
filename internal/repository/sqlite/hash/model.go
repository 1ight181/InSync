package hash

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type HashCacheEntry struct {
	Id       string `gorm:"primaryKey;type:uuid"`
	Hash     string `gorm:"type:varchar(255);not null"`
	FullPath string `gorm:"type:varchar(255);not null;uniqueIndex:ux_hash_cache_full_path"`
	Expires  int64  `gorm:"not null"`
}

func (e *HashCacheEntry) BeforeCreate(tx *gorm.DB) error {
	e.Id = uuid.NewString()
	return nil
}
