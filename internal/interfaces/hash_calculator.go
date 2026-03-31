package interfaces

import "insync/internal/domain"

type IHashCalculator interface {
	CalculateHash(content domain.ResourceContent) (string, error)
}
