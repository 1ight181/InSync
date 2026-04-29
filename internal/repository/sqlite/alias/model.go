package alias

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Alias struct {
	Id       string `gorm:"primaryKey;type:uuid"`
	Name     string `gorm:"type:varchar(255);not null"`
	NodeName string `gorm:"type:varchar(255);not null;uniqueIndex"`
}

func (a *Alias) BeforeCreate(tx *gorm.DB) error {
	a.Id = uuid.NewString()
	return nil
}
