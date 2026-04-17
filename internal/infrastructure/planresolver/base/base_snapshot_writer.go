package base

import (
	"context"
	"insync/internal/domain"
)

type BaseSnapshotWriter struct {
	baseSnapshotRepository IBaseSnapshotRepositoryWriter
}

type BaseSnapshotWriterOptions struct {
	BaseSnapshotRepository IBaseSnapshotRepositoryWriter
}

func NewBaseSnapshotWriter(opts BaseSnapshotWriterOptions) *BaseSnapshotWriter {
	if opts.BaseSnapshotRepository == nil {
		panic("Все поля BaseSnapshotWriterOptions должны быть заполнены")
	}
	return &BaseSnapshotWriter{baseSnapshotRepository: opts.BaseSnapshotRepository}
}

func (b *BaseSnapshotWriter) CreateBaseSnapshot(ctx context.Context, baseSnapshot domain.Snapshot) error {
	return b.baseSnapshotRepository.CreateBaseSnapshot(ctx, baseSnapshot)
}
