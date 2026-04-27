package syncer

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type Syncer struct {
	changeApplier                 IChangeApplier
	conflictResolver              IConflictResolver
	postSyncBaseSnapshotPersister IPostSyncBaseSnapshotPersister
}

type SyncerOptions struct {
	ChangeApplier                 IChangeApplier
	ConflictResolver              IConflictResolver
	PostSyncBaseSnapshotPersister IPostSyncBaseSnapshotPersister
}

var (
	ErrInvalidSyncerOptions = errors.New("Все поля SyncerOptions должны быть заполнены")
)

func NewSyncer(opts SyncerOptions) (*Syncer, error) {
	if opts.ChangeApplier == nil ||
		opts.ConflictResolver == nil ||
		opts.PostSyncBaseSnapshotPersister == nil {
		return nil, ErrInvalidSyncerOptions
	}
	return &Syncer{
		changeApplier:                 opts.ChangeApplier,
		conflictResolver:              opts.ConflictResolver,
		postSyncBaseSnapshotPersister: opts.PostSyncBaseSnapshotPersister,
	}, nil
}

func (s *Syncer) Sync(ctx context.Context, plan domain.SyncPlan, rootName domain.RootName) (
	<-chan domain.ChangeEvent,
	<-chan domain.Conflict,
	<-chan error,
	chan<- domain.Decision,
	error,
) {
	appliedChanges := make(chan domain.ChangeEvent, plan.ChangeLength())
	conflicts := make(chan domain.Conflict, plan.ConflictLength())
	baseSnapshotSaveError := make(chan error, 1)
	userDecision := make(chan domain.Decision, 1)

	go func() {
		for _, change := range plan.LocalChanges {
			if err := s.applyLocalChange(ctx, rootName, change); err != nil {
				appliedChanges <- domain.ChangeEvent{Err: err}
				continue
			}

			appliedChanges <- domain.ChangeEvent{Change: change.ToSyncChange()}
		}

		for _, change := range plan.RemoteChanges {
			if err := s.applyRemoteChange(ctx, rootName, change); err != nil {
				appliedChanges <- domain.ChangeEvent{Err: err}
				continue
			}

			appliedChanges <- domain.ChangeEvent{Change: change.ToSyncChange()}
		}

		for _, conflict := range plan.Conflicts {
			conflicts <- conflict
			select {
			case <-ctx.Done():
				return
			case decision := <-userDecision:
				appliedChange, err := s.resolveConflict(ctx, rootName, conflict, decision)
				if err != nil {
					if errors.Is(err, ErrShouldSkip) {
						continue
					}

					appliedChanges <- domain.ChangeEvent{Err: err}
					continue
				}

				appliedChanges <- domain.ChangeEvent{Change: appliedChange}
			}
		}

		if err := s.postSyncBaseSnapshotPersister.UpdateBaseSnapshot(ctx, rootName); err != nil {
			baseSnapshotSaveError <- err
		}

		close(appliedChanges)
		close(conflicts)
		close(userDecision)
		close(baseSnapshotSaveError)

	}()

	return appliedChanges, conflicts, baseSnapshotSaveError, userDecision, nil
}

func (s *Syncer) applyLocalChange(
	ctx context.Context,
	rootName domain.RootName, change domain.LocalChange,
) error {
	if err := s.changeApplier.ApplyLocal(ctx, rootName, change); err != nil {
		return err
	}

	return nil
}

func (s *Syncer) applyRemoteChange(
	ctx context.Context,
	rootName domain.RootName, change domain.RemoteChange,
) error {
	if err := s.changeApplier.ApplyRemote(ctx, rootName, change); err != nil {
		return err
	}
	return nil
}

func (s *Syncer) resolveConflict(
	ctx context.Context,
	rootName domain.RootName,
	conflict domain.Conflict,
	userDecision domain.Decision,
) (domain.SyncChange, error) {
	requiredChange, isLocal, err := s.conflictResolver.Resolve(conflict, userDecision)
	if err != nil {
		return domain.SyncChange{}, err
	}

	if isLocal {
		if err := s.changeApplier.ApplyLocal(ctx, rootName, requiredChange.ToLocalChange()); err != nil {
			return domain.SyncChange{}, err
		}
	} else {
		if err := s.changeApplier.ApplyRemote(ctx, rootName, requiredChange.ToRemoteChange()); err != nil {
			return domain.SyncChange{}, err
		}
	}

	return domain.SyncChange{}, nil
}
