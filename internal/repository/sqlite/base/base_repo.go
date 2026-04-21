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

var (
	ErrInvalidBaseSnapshotRepositoryOptions = errors.New("Все поля BaseSnapshotRepositoryOptions должны быть заполнены")
)

func NewBaseSnapshotRepository(opts BaseSnapshotRepositoryOptions) (*BaseSnapshotRepository, error) {
	if opts.Db == nil {
		return nil, ErrInvalidBaseSnapshotRepositoryOptions
	}
	return &BaseSnapshotRepository{db: opts.Db}, nil
}

func (b *BaseSnapshotRepository) GetLastBaseSnapshotByDeviceIdAndRootName(
	ctx context.Context,
	localDeviceId, remoteDeviceId domain.DeviceId,
	rootName domain.RootName,
) (
	domain.Snapshot, error,
) {
	var baseSnapshot BaseSnapshot

	err := b.db.
		WithContext(ctx).
		Preload("Files.FileInfo.FileMetadata").
		Order("created_at desc").
		First(&baseSnapshot,
			"local_device_id = ? AND remote_device_id = ? AND root_name = ?",
			localDeviceId.String(), remoteDeviceId.String(), rootName.String()).
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

func (b *BaseSnapshotRepository) CreateBaseSnapshot(ctx context.Context,
	baseSnapshot domain.Snapshot,
	localDeviceId, remoteDeviceId domain.DeviceId,
	rootName domain.RootName,
) error {
	baseSnapshotModel := ToBaseSnapshot(baseSnapshot, localDeviceId, remoteDeviceId, rootName)
	return b.db.WithContext(ctx).Create(&baseSnapshotModel).Error
}
