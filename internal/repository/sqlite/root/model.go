package root

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Root struct {
	Id       string `gorm:"primaryKey;type:uuid"`
	RootName string `gorm:"type:varchar(255);not null;uniqueIndex"`
	RootPath string `gorm:"type:varchar(255);not null;uniqueIndex"`
}

func (r *Root) BeforeCreate(tx *gorm.DB) error {
	r.Id = uuid.NewString()
	return nil
}
