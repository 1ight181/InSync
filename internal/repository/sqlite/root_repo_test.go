package sqlite

import (
	"insync/internal/domain"
	"log/slog"
	"testing"

	root "insync/internal/repository/sqlite/root"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type RootRepositorySuite struct {
	suite.Suite

	db   *gorm.DB
	repo *root.RootRepository
}

func TestRootRepositorySuite(t *testing.T) {
	suite.Run(t, new(RootRepositorySuite))
}

func (s *RootRepositorySuite) SetupTest() {
	migrator := NewAutoMigrator()

	dbOpts := SqliteGormOptions{
		Dsn:      ":memory:",
		Migrator: migrator,
		Logger:   slog.Default(),
	}

	db, err := NewGorm(dbOpts)
	s.Require().NoError(err)

	s.db = db
	s.repo = root.NewRootRepository(db)
}

func (s *RootRepositorySuite) TearDownSuite() {
	db, _ := s.db.DB()
	db.Close()
}

func (s *RootRepositorySuite) TestRootRepository_GetRoots_Empty() {
	rootMap, err := s.repo.GetRoots()

	s.Require().NoError(err)
	s.Require().Empty(rootMap)
}

func (s *RootRepositorySuite) TestRootRepository_AddRoot_And_GetRoots() {
	rootName := domain.RootName("root1")

	path, err := domain.NewPath("path/to/root1")
	s.Require().NoError(err)

	err = s.repo.AddRoot(rootName, path)
	s.Require().NoError(err)

	rootMap, err := s.repo.GetRoots()
	s.Require().NoError(err)

	s.Require().Len(rootMap, 1)
	s.Require().Equal(path, rootMap[rootName])
}

func (s *RootRepositorySuite) TestRootRepository_AddRoot_Overwrite() {
	rootName := domain.RootName("root1")

	path1, err := domain.NewPath("path/one")
	s.Require().NoError(err)

	path2, err := domain.NewPath("path/two")
	s.Require().NoError(err)

	err = s.repo.AddRoot(rootName, path1)
	s.Require().NoError(err)

	err = s.repo.AddRoot(rootName, path2)
	s.Require().NoError(err)

	rootMap, err := s.repo.GetRoots()
	s.Require().NoError(err)

	s.Require().Len(rootMap, 1)
	s.Require().Equal(path2, rootMap[rootName])
}

func (s *RootRepositorySuite) TestRootRepository_RemoveRoot() {
	rootName := domain.RootName("root1")

	path, err := domain.NewPath("path/to/root")
	s.Require().NoError(err)

	err = s.repo.AddRoot(rootName, path)
	s.Require().NoError(err)

	err = s.repo.RemoveRoot(rootName)
	s.Require().NoError(err)

	rootMap, err := s.repo.GetRoots()
	s.Require().NoError(err)

	s.Require().Empty(rootMap)
}

func (s *RootRepositorySuite) TestRootRepository_RemoveRoot_NonExisting() {
	err := s.repo.RemoveRoot(domain.RootName("missing"))
	s.Require().NoError(err)
}

func (s *RootRepositorySuite) TestRootRepository_GetRoots_DbError() {
	db, _ := s.db.DB()
	db.Close()

	_, err := s.repo.GetRoots()

	s.Require().Error(err)
}
