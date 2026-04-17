package repository

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SqliteGormOptions struct {
	Dsn      string
	Migrator IMigrator
}

func NewGorm(opts SqliteGormOptions) *gorm.DB {
	if opts.Dsn == "" ||
		opts.Migrator == nil {
		panic("Все поля SqliteGormOptions должны быть заполнены")
	}

	db, err := gorm.Open(sqlite.Open(opts.Dsn))
	if err != nil {
		panic(err)
	}

	if err := opts.Migrator.Migrate(db); err != nil {
		panic(err)
	}

	return db
}
