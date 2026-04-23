package sqlite

import (
	"errors"
	"log/slog"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SqliteGormOptions struct {
	Dsn      string
	Migrator IMigrator
	Logger   *slog.Logger
}

var (
	ErrInvalidSqliteGormOptions = errors.New("Все поля SqliteGormOptions должны быть заполнены")
)

func NewGorm(opts SqliteGormOptions) (*gorm.DB, error) {
	if opts.Dsn == "" ||
		opts.Migrator == nil ||
		opts.Logger == nil {
		return nil, ErrInvalidSqliteGormOptions
	}

	db, err := gorm.Open(sqlite.Open(opts.Dsn), &gorm.Config{
		Logger: logger.NewSlogLogger(opts.Logger, logger.Config{
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		}),
		TranslateError: true,
	})
	if err != nil {
		return nil, err
	}

	if err := opts.Migrator.Migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}
