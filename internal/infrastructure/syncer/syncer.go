package syncer

import (
	"context"
	"insync/internal/domain"
)

type Syncer struct {
	changeApplier    IChangeApplier
	conflictResolver IConflictResolver
}

type SyncerOptions struct {
	ChangeApplier    IChangeApplier
	ConflictResolver IConflictResolver
}

func NewSyncer(opts SyncerOptions) *Syncer {
	if opts.ChangeApplier == nil ||
		opts.ConflictResolver == nil {
		panic("Все поля SyncerOptions должны быть заполнены")
	}
	return &Syncer{
		changeApplier:    opts.ChangeApplier,
		conflictResolver: opts.ConflictResolver,
	}
}

func (s *Syncer) Sync(ctx context.Context, plan domain.SyncPlan, rootName domain.RootName) (
	<-chan domain.ChangeEvent,
	<-chan domain.Conflict,
	chan<- domain.Decision,
	error,
) {
	appliedChanges := make(chan domain.ChangeEvent, plan.ChangeLength())
	conflicts := make(chan domain.Conflict, plan.ConflictLength())
	userDecision := make(chan domain.Decision, 1)

	go func() {
		for _, change := range plan.LocalChanges {
			if err := s.applyLocalChange(ctx, rootName, change); err != nil {
				appliedChanges <- domain.ChangeEvent{Err: err}
			}

			appliedChanges <- domain.ChangeEvent{Change: change.ToSyncChange()}
		}

		for _, change := range plan.RemoteChanges {
			if err := s.applyRemoteChange(ctx, rootName, change); err != nil {
				appliedChanges <- domain.ChangeEvent{Err: err}
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
					appliedChanges <- domain.ChangeEvent{Err: err}
				}

				appliedChanges <- domain.ChangeEvent{Change: appliedChange}
			}
		}

		close(appliedChanges)
		close(conflicts)
		close(userDecision)
	}()

	return appliedChanges, conflicts, userDecision, nil
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
