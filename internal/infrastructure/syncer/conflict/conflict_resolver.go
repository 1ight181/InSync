package conflict

import (
	"errors"
	"insync/internal/domain"
)

var (
	ErrUnknownConflictType = errors.New("неизвестный тип конфликта")
)

type ConflictResolver struct{}

func NewConflictResolver() *ConflictResolver {
	return &ConflictResolver{}
}

func (c *ConflictResolver) Resolve(conflict domain.Conflict, userDecision domain.Decision) (domain.SyncChange, bool, error) {
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
