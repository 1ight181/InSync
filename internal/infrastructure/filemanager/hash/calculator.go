package hash

import (
	"encoding/hex"
	"errors"
	"io"

	cont "insync/internal/infrastructure/filemanager/content"

	"github.com/zeebo/blake3"
)

type IHashCalculator interface {
	CalculateHash(content cont.ResourceContent) (string, error)
}

type HashCalculator struct{}

func NewHashCalculator() IHashCalculator {
	return &HashCalculator{}
}

func (hc *HashCalculator) CalculateHash(content cont.ResourceContent) (string, error) {
	hasher := blake3.New()

	contentReader, err := content.OpenContent(content.FullPath, content.RelativePath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = contentReader.Close()
	}()

	buf := make([]byte, 4096)

	for {
		n, err := contentReader.Read(buf)
		if n > 0 {
			if _, wErr := hasher.Write(buf[:n]); wErr != nil {
				return "", wErr
			}
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", err
		}
	}

	sum := hasher.Sum(nil)
	return hex.EncodeToString(sum), nil
}
