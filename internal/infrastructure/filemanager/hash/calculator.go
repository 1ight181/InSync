package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	cont "insync/internal/infrastructure/filemanager/content"
	"io"
)

type IHashCalculator interface {
	CalculateHash(content cont.ResourceContent) (string, error)
}

type HashCalculator struct{}

func NewHashCalculator() IHashCalculator {
	return &HashCalculator{}
}

func (hc *HashCalculator) CalculateHash(content cont.ResourceContent) (string, error) {
	hash := sha256.New()
	fullPath := content.FullPath
	relativePath := content.RelativePath
	contentReader, err := content.OpenContent(fullPath, relativePath)
	if err != nil {
		return "", err
	}
	for {
		buf := make([]byte, 4096)
		n, err := contentReader.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", err
		}
		_, err = hash.Write(buf[:n])
		if err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
