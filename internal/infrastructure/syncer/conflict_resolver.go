package syncer

import (
	"errors"
	"insync/internal/domain"
)

type IConflictResolver interface {
	Resolve(conflicts domain.Conflict, userDecision domain.Decision) (requiredChange domain.SyncChange, isLocal bool, err error)
}

var (
	ErrUnknownConflictType = errors.New("неизвестный тип конфликта")
	ErrShouldSkip          = errors.New("пользователь пропустил конфликт")
)

type ConflictResolver struct{}

func NewConflictResolver() *ConflictResolver {
	return &ConflictResolver{}
}

func (c *ConflictResolver) Resolve(conflict domain.Conflict, userDecision domain.Decision) (domain.SyncChange, bool, error) {
	if userDecision.Equal(domain.Skip) {
		return domain.SyncChange{}, false, ErrShouldSkip
	}

	switch conflict.Conflict {
	case domain.ConflictLocalDeletedRemoteModified:
		if userDecision.Equal(domain.LocalWin) {
			return domain.SyncChange{
				ChangeType:      domain.Delete,
				OldRelativePath: conflict.RemoteRelativePath,
			}, false, nil
		}

		return domain.SyncChange{
			ChangeType:      domain.Create,
			NewRelativePath: conflict.RemoteRelativePath,
		}, true, nil

	case domain.ConflictRemoteDeletedLocalModified:
		if userDecision.Equal(domain.LocalWin) {
			return domain.SyncChange{
				ChangeType:      domain.Create,
				NewRelativePath: conflict.LocalRelativePath,
			}, false, nil
		}

		return domain.SyncChange{
			ChangeType:      domain.Delete,
			OldRelativePath: conflict.LocalRelativePath,
		}, true, nil

	case domain.ConflictLocalMovedRemoteMoved:
		if userDecision.Equal(domain.LocalWin) {
			return domain.SyncChange{
				ChangeType:      domain.Move,
				OldRelativePath: conflict.RemoteRelativePath,
				NewRelativePath: conflict.LocalRelativePath,
			}, false, nil
		}

		return domain.SyncChange{
			ChangeType:      domain.Move,
			OldRelativePath: conflict.LocalRelativePath,
			NewRelativePath: conflict.RemoteRelativePath,
		}, true, nil

	case domain.ConflictLocalRenamedRemoteRenamed:
		if userDecision.Equal(domain.LocalWin) {
			return domain.SyncChange{
				ChangeType:      domain.Rename,
				OldRelativePath: conflict.RemoteRelativePath,
				NewRelativePath: conflict.LocalRelativePath,
			}, false, nil
		}

		return domain.SyncChange{
			ChangeType:      domain.Rename,
			OldRelativePath: conflict.LocalRelativePath,
			NewRelativePath: conflict.RemoteRelativePath,
		}, true, nil

	case domain.ConflictBothModifiedAtSameTime, domain.ConflictBothCreatedAtSamePathConflict:
		if userDecision.Equal(domain.LocalWin) {
			return domain.SyncChange{
				ChangeType:      domain.Modify,
				OldRelativePath: conflict.RemoteRelativePath,
			}, false, nil
		}

		return domain.SyncChange{
			ChangeType:      domain.Modify,
			OldRelativePath: conflict.LocalRelativePath,
		}, true, nil
	default:
		return domain.SyncChange{}, false, ErrUnknownConflictType
	}

}
