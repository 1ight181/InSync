package dirty

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DirtyPaths struct {
	Id       string `gorm:"primaryKey;type:uuid"`
	FullPath string `gorm:"type:varchar(255);not null;uniqueIndex:ux_dirty_full_path"`
}

func (d *DirtyPaths) BeforeCreate(tx *gorm.DB) error {
	d.Id = uuid.NewString()
	return nil
}
