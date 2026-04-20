package syncer

import "insync/internal/domain"

type ILocalApplier interface {
	Apply(change domain.LocalChange) error
}
