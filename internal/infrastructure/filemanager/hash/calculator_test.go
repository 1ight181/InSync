package hash

import (
	"bytes"
	"io"
	"testing"

	"insync/internal/domain"
	cont "insync/internal/infrastructure/filemanager/content"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type HashCalculatorSuite struct {
	suite.Suite

	calculator IHashCalculator
}

func (s *HashCalculatorSuite) SetupTest() {
	s.calculator = NewHashCalculator()
}

func TestHashCalculator(t *testing.T) {
	suite.Run(t, new(HashCalculatorSuite))
}

func (s *HashCalculatorSuite) TestCalculateHash_SameContent_SameResult() {
	content := []byte("same-content-data")

	resourceA := cont.ResourceContent{
		FullPath:     mustPath(s.T(), "/path/to/file.txt"),
		RelativePath: mustPath(s.T(), "file.txt"),
		OpenContent: func(fullPath domain.Path, relativePath domain.Path) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(content)), nil
		},
	}

	resourceB := cont.ResourceContent{
		FullPath:     mustPath(s.T(), "/other/path/to/file.txt"),
		RelativePath: mustPath(s.T(), "file.txt"),
		OpenContent: func(fullPath domain.Path, relativePath domain.Path) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(content)), nil
		},
	}

	hashA, err := s.calculator.CalculateHash(resourceA)
	s.Require().NoError(err)
	hashB, err := s.calculator.CalculateHash(resourceB)
	s.Require().NoError(err)

	s.Require().NotEmpty(hashA)
	s.Require().Equal(hashA, hashB)
}

func (s *HashCalculatorSuite) TestCalculateHash_DifferentContent_DifferentResult() {
	resourceA := cont.ResourceContent{
		FullPath:     mustPath(s.T(), "/path/to/file.txt"),
		RelativePath: mustPath(s.T(), "file.txt"),
		OpenContent: func(fullPath domain.Path, relativePath domain.Path) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader([]byte("content-a"))), nil
		},
	}

	resourceB := cont.ResourceContent{
		FullPath:     mustPath(s.T(), "/path/to/file.txt"),
		RelativePath: mustPath(s.T(), "file.txt"),
		OpenContent: func(fullPath domain.Path, relativePath domain.Path) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader([]byte("content-b"))), nil
		},
	}

	hashA, err := s.calculator.CalculateHash(resourceA)
	s.Require().NoError(err)
	hashB, err := s.calculator.CalculateHash(resourceB)
	s.Require().NoError(err)

	s.Require().NotEqual(hashA, hashB)
}

func (s *HashCalculatorSuite) TestCalculateHash_OpenContentError_ReturnsError() {
	resource := cont.ResourceContent{
		FullPath:     mustPath(s.T(), "/path/to/file.txt"),
		RelativePath: mustPath(s.T(), "file.txt"),
		OpenContent: func(fullPath domain.Path, relativePath domain.Path) (io.ReadCloser, error) {
			return nil, assertError{}
		},
	}

	_, err := s.calculator.CalculateHash(resource)
	s.Require().Error(err)
}

type assertError struct{}

func (assertError) Error() string {
	return "open content failed"
}

func mustPath(t *testing.T, rawPath string) domain.Path {
	t.Helper()
	p, err := domain.NewPath(rawPath)
	require.NoError(t, err)
	return p
}
