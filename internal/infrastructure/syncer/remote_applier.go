package syncer

import "insync/internal/domain"

type IRemoteApplier interface {
	Apply(change domain.RemoteChange) error
}
