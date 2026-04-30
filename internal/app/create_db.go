package app

import (
	"log/slog"
	"os"

	config "insync/internal/infrastructure/config/models"
	sql "insync/internal/repository/sqlite"

	"gorm.io/gorm"
)

func createDb(
	dbConfig config.DbConfig,
	dbLogger *slog.Logger,
	cleanups *cleanupStack,
) (*gorm.DB, error) {
	if err := os.MkdirAll(dbConfig.Dir, 0755); err != nil {
		return nil, err
	}

	migrator := sql.NewAutoMigrator()
	dbOpts := sql.SqliteGormOptions{
		Dsn:      dbConfig.GetDsn(),
		Migrator: migrator,
		Logger:   dbLogger,
	}

	db, err := sql.NewGorm(dbOpts)
	if err != nil {
		return nil, err
	}

	cleanups.Add(func() {
		sqlDb, err := db.DB()
		if err != nil {
			dbLogger.Warn("Не удалось получить sqlDb для закрытия sqlite")
		}
		if err := sqlDb.Close(); err != nil {
			dbLogger.Warn("Не удалось коректно закрыть sqlite")
		}

	})

	return db, nil
}
