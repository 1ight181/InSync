package interfaces

import (
	models "insync/internal/mdns/models"
)

type IResolver interface {
	Browse() (chan models.ResolvedAddresses, error)
}
