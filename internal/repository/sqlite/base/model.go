package base

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseSnapshot struct {
	Id             string `gorm:"primaryKey;type:uuid"`
	IsInitial      bool
	RootName       string `gorm:"type:uuid;not null;index:idx_base_snapshot_root_name_local_device_id_remote_device_id"`
	LocalDeviceId  string `gorm:"type:uuid;not null;index:idx_base_snapshot_root_name_local_device_id_remote_device_id"`
	RemoteDeviceId string `gorm:"type:uuid;not null;index:idx_base_snapshot_root_name_local_device_id_remote_device_id"`

	CreatedAt int64 `gorm:"autoCreateTime"`

	Files []FileEntry `gorm:"foreignKey:BaseSnapshotId;references:Id;onDelete:CASCADE"`
}

func (b *BaseSnapshot) BeforeCreate(tx *gorm.DB) error {
	b.Id = uuid.NewString()
	return nil
}

type FileEntry struct {
	Id             string `gorm:"primaryKey;type:uuid"`
	BaseSnapshotId string `gorm:"type:uuid"`

	RelativePath string `gorm:"type:varchar(255);not null"`
	SubtreeSize  uint64 `gorm:"not null"`

	CreatedAt int64 `gorm:"autoCreateTime"`

	FileInfo FileInfo `gorm:"foreignKey:FileEntryId;references:Id;onDelete:CASCADE"`
}

func (e *FileEntry) BeforeCreate(tx *gorm.DB) error {
	e.Id = uuid.NewString()
	return nil
}

type FileInfo struct {
	Id          string `gorm:"primaryKey;type:uuid"`
	FileEntryId string `gorm:"type:uuid;uniqueIndex"`

	Hash string `gorm:"type:varchar(255);not null"`

	CreatedAt int64 `gorm:"autoCreateTime"`

	FileMetadata FileMetadata `gorm:"foreignKey:FileInfoId;references:Id;onDelete:CASCADE"`
}

func (f *FileInfo) BeforeCreate(tx *gorm.DB) error {
	f.Id = uuid.NewString()
	return nil
}

type FileMetadata struct {
	Id         string `gorm:"primaryKey;type:uuid"`
	FileInfoId string `gorm:"type:uuid;uniqueIndex"`

	IsDirectory  bool
	SizeBytes    uint64
	ModifiedUnix uint64

	CreatedAt int64 `gorm:"autoCreateTime"`
}

func (m *FileMetadata) BeforeCreate(tx *gorm.DB) error {
	m.Id = uuid.NewString()
	return nil
}
