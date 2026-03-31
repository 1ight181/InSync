package interfaces

import "insync/internal/domain"

type IHashCalculator interface {
	CalculateHash(content domain.ResourceContent) (string, error)
	CalculateAggregateHash(contents []domain.ResourceContent) (string, error)
}
