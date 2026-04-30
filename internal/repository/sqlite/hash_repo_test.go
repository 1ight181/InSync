package sqlite

import (
	"insync/internal/domain"
	hashrepo "insync/internal/repository/sqlite/hash"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type HashRepositorySuite struct {
	suite.Suite

	db   *gorm.DB
	repo *hashrepo.HashRepository
}

func (s *HashRepositorySuite) SetupTest() {
	migrator := NewAutoMigrator()
	dbOpts := SqliteGormOptions{
		Migrator: migrator,
		Dsn:      ":memory:",
		Logger:   slog.Default(),
	}

	db, err := NewGorm(dbOpts)
	s.Require().NoError(err)
	s.db = db

	opts := hashrepo.HashRepositoryOptions{
		Db:                           db,
		HashCacheEntryExpireUnixTime: 10,
	}

	repo, err := hashrepo.NewHashRepository(opts)
	s.Require().NoError(err)
	s.repo = repo
}

func (s *HashRepositorySuite) TearDownSuite() {
	db, _ := s.db.DB()
	db.Close()
}

func TestHashRepositorySuite(t *testing.T) {
	suite.Run(t, new(HashRepositorySuite))
}

func (s *HashRepositorySuite) TestHashRepositorySuite_Get_Empty() {
	hashCache, err := s.repo.GetHashCache()
	s.Require().NoError(err)
	s.Require().Empty(hashCache)
}

func (s *HashRepositorySuite) TestHashRepositorySuite_SetHashCache_Success() {
	path1 := domain.Path("path")
	hash1 := "hash"
	err := s.repo.SetHashCache(path1, hash1)
	s.Require().NoError(err)

	path2 := domain.Path("path/to")
	hash2 := "hash1"
	err = s.repo.SetHashCache(path2, hash2)
	s.Require().NoError(err)

	path3 := domain.Path("path/to/file2")
	hash3 := "hash2"
	err = s.repo.SetHashCache(path3, hash3)
	s.Require().NoError(err)

	expected := map[domain.Path]string{
		path1: hash1,
		path2: hash2,
		path3: hash3,
	}

	hashCache, err := s.repo.GetHashCache()
	s.Require().NoError(err)
	s.Require().Equal(expected, hashCache)

	for fullPath, hash := range hashCache {
		s.Require().Equal(expected[fullPath], hash)
	}
}
