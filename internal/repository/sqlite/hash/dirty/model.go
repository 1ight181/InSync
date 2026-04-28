package dirty

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DirtyPath struct {
	Id       string `gorm:"primaryKey;type:uuid"`
	FullPath string `gorm:"type:varchar(255);not null;uniqueIndex:ux_dirty_full_path"`
	RootName string `gorm:"type:varchar(255);not null;uniqueIndex:ux_dirty_root_name"`
}

func (d *DirtyPath) BeforeCreate(tx *gorm.DB) error {
	d.Id = uuid.NewString()
	return nil
}
