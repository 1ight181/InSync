package interfaces

import (
	"insync/internal/domain"
)

type IMdnsBrowser interface {
	Browse() (chan domain.Node, error)
}
