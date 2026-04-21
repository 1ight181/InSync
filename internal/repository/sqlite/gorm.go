package sqlite

import (
	"errors"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SqliteGormOptions struct {
	Dsn      string
	Migrator IMigrator
}

var (
	ErrInvalidSqliteGormOptions = errors.New("Все поля SqliteGormOptions должны быть заполнены")
)

func NewGorm(opts SqliteGormOptions) (*gorm.DB, error) {
	if opts.Dsn == "" ||
		opts.Migrator == nil {
		return nil, ErrInvalidSqliteGormOptions
	}

	db, err := gorm.Open(sqlite.Open(opts.Dsn))
	if err != nil {
		return nil, err
	}

	if err := opts.Migrator.Migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}
