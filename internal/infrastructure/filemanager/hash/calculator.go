package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"insync/internal/domain"
	"insync/internal/interfaces"
)

type HashCalculator struct{}

func NewHashCalculator() interfaces.IHashCalculator {
	return &HashCalculator{}
}

func (hc *HashCalculator) CalculateHash(content domain.ResourceContent) (string, error) {
	hash := sha256.New()
	_, err := hash.Write(content.Content)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (hc *HashCalculator) CalculateAggregateHash(contents []domain.ResourceContent) (string, error) {
	hash := sha256.New()
	for _, content := range contents {
		_, err := hash.Write(content.Content)
		if err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
