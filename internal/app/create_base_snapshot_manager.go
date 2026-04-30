package app

import (
	"insync/internal/infrastructure/base"
	"insync/internal/infrastructure/deviceid"
	baserepo "insync/internal/repository/sqlite/base"

	"gorm.io/gorm"
)

func createBaseSnapshotManager(
	deviceIdProvider *deviceid.DeviceIdProvider,
	db *gorm.DB,
) (*base.BaseSnapshotManager, error) {
	baseSnapshotRepository, err := baserepo.NewBaseSnapshotRepository(db)
	if err != nil {
		return nil, err
	}

	baseSnapshotManagerOpts := base.BaseSnapshotManagerOptions{
		BaseSnapshotRepository: baseSnapshotRepository,
		DeviceIdProvider:       deviceIdProvider,
	}

	baseSnapshotManager, err := base.NewBaseSnapshotManager(baseSnapshotManagerOpts)
	if err != nil {
		return nil, err
	}

	return baseSnapshotManager, nil
}
