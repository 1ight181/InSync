package base

import (
	"context"
	"errors"
	"insync/internal/domain"

	"gorm.io/gorm"
)

type BaseSnapshotRepository struct {
	db *gorm.DB
}

type BaseSnapshotRepositoryOptions struct {
	Db *gorm.DB
}

func NewBaseSnapshotRepository(opts BaseSnapshotRepositoryOptions) *BaseSnapshotRepository {
	if opts.Db == nil {
		panic("все поля BaseSnapshotRepositoryOptions должны быть заполнены")
	}
	return &BaseSnapshotRepository{db: opts.Db}
}

func (b *BaseSnapshotRepository) GetLastBaseSnapshotByDeviceIdAndRootName(
	ctx context.Context,
	localDeviceId, remoteDeviceId string,
	rootName domain.RootName,
) (
	domain.Snapshot, error,
) {
	var baseSnapshot BaseSnapshot

	err := b.db.
		WithContext(ctx).
		Preload("Files.FileEntry.FileInfo.FileMetadata").
		Order("unix_time desc").
		First(&baseSnapshot,
			"local_device_id = ? AND remote_device_id = ? AND root_name = ?",
			localDeviceId, remoteDeviceId, rootName).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Snapshot{}, ErrBaseSnapshotNotFound
		}

		return domain.Snapshot{}, err
	}

	domainBaseSnapshot := ToDomainSnapshot(baseSnapshot)

	return domainBaseSnapshot, nil
}

func (b *BaseSnapshotRepository) CreateBaseSnapshot(ctx context.Context, baseSnapshot domain.SnapshotWithMetadata) error {
	baseSnapshotModel := ToBaseSnapshot(baseSnapshot)
	return b.db.WithContext(ctx).Create(baseSnapshotModel).Error
}
