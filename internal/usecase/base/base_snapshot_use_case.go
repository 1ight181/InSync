package base

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type BaseSnapshotUseCase struct {
	baseSnapshotCreator IBaseSnapshotCreator
}

type BaseSnapshotUseCaseOpts struct {
	BaseSnapshotCreator IBaseSnapshotCreator
}

var (
	ErrInvalidOpts = errors.New("Все поля BaseSnapshotUseCaseOptions должны быть заполнены")
)

func NewBaseSnapshotUseCase(opts BaseSnapshotUseCaseOpts) (*BaseSnapshotUseCase, error) {
	if opts.BaseSnapshotCreator == nil {
		return nil, ErrInvalidOpts
	}
	return &BaseSnapshotUseCase{baseSnapshotCreator: opts.BaseSnapshotCreator}, nil
}

func (b *BaseSnapshotUseCase) SetBaseSnapshot(ctx context.Context, baseSnapshot domain.Snapshot, rootName domain.RootName) error {
	return b.baseSnapshotCreator.CreateBaseSnapshot(ctx, baseSnapshot, rootName)
}
