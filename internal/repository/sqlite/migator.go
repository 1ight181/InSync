package repository

import (
	snap "insync/internal/repository/sqlite/base"
	hash "insync/internal/repository/sqlite/hash"

	"gorm.io/gorm"
)

type IMigrator interface {
	Migrate(db *gorm.DB) error
}

type AutoMigrator struct{}

func NewAutoMigrator() *AutoMigrator {
	return &AutoMigrator{}
}

func (a *AutoMigrator) Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&hash.HashCacheEntry{},
		&snap.BaseSnapshot{},
		&snap.FileEntry{},
		&snap.FileInfo{},
		&snap.FileMetadata{},
	)
	if err != nil {
		return err
	}

	return nil
}
