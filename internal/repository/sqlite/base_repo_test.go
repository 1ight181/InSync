package sqlite

import (
	"context"
	"insync/internal/domain"
	baserepo "insync/internal/repository/sqlite/base"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type BaseRepositorySuite struct {
	suite.Suite
	repo *baserepo.BaseSnapshotRepository
	db   *gorm.DB
}

func (s *BaseRepositorySuite) SetupSuite() {
	migrator := NewAutoMigrator()
	dbOpts := SqliteGormOptions{
		Dsn:      ":memory:",
		Migrator: migrator,
	}

	db, err := NewGorm(dbOpts)
	s.Require().NoError(err)
	s.db = db

	repo, err := baserepo.NewBaseSnapshotRepository(baserepo.BaseSnapshotRepositoryOptions{Db: db})
	s.Require().NoError(err)
	s.repo = repo
}

func TestBaseRepositorySuite(t *testing.T) {
	suite.Run(t, new(BaseRepositorySuite))
}

func (s *BaseRepositorySuite) TearDownSuite() {
	db, _ := s.db.DB()
	db.Close()
}

func (s *BaseRepositorySuite) TestBaseRepository_Get_Empty() {

	snapshot, err := s.repo.GetLastBaseSnapshotByDeviceIdAndRootName(context.Background(), "", "", "")
	s.Require().ErrorIs(err, baserepo.ErrBaseSnapshotNotFound)
	s.Require().Equal(domain.Snapshot{}, snapshot)
}

func (s *BaseRepositorySuite) TestBaseRepository_Create() {
	ctx := context.Background()
	baseSnapshot := s.getDefaultSnapshot()
	localDeviceId := domain.DeviceId("device1")
	remoteDeviceId := domain.DeviceId("device2")
	rootName := domain.RootName("root")
	err := s.repo.CreateBaseSnapshot(ctx, baseSnapshot, localDeviceId, remoteDeviceId, rootName)
	s.Require().NoError(err)

	snapshot, err := s.repo.GetLastBaseSnapshotByDeviceIdAndRootName(ctx, localDeviceId, remoteDeviceId, rootName)
	s.Require().NoError(err)
	s.Require().Equal(baseSnapshot, snapshot)
}

func (s *BaseRepositorySuite) TestBaseRepository_Get_GetLastFromMany() {
	ctx := context.Background()
	baseSnapshot1 := s.getDefaultSnapshot()
	localDeviceId1 := domain.DeviceId("device1")
	remoteDeviceId1 := domain.DeviceId("device2")
	rootName1 := domain.RootName("root")
	err := s.repo.CreateBaseSnapshot(ctx, baseSnapshot1, localDeviceId1, remoteDeviceId1, rootName1)
	s.Require().NoError(err)

	time.Sleep(time.Millisecond)

	baseSnapshot2 := s.getDefaultSnapshot()
	localDeviceId2 := domain.DeviceId("device3")
	remoteDeviceId2 := domain.DeviceId("device4")
	rootName2 := domain.RootName("root")
	err = s.repo.CreateBaseSnapshot(ctx, baseSnapshot2, localDeviceId2, remoteDeviceId2, rootName2)
	s.Require().NoError(err)

	time.Sleep(time.Millisecond)

	baseSnapshot3 := s.getDefaultSnapshot()
	localDeviceId3 := domain.DeviceId("device1")
	remoteDeviceId3 := domain.DeviceId("device2")
	rootName3 := domain.RootName("lab")
	err = s.repo.CreateBaseSnapshot(ctx, baseSnapshot3, localDeviceId3, remoteDeviceId3, rootName3)
	s.Require().NoError(err)

	time.Sleep(time.Millisecond)

	baseSnapshot4 := s.getDefaultSnapshot()
	localDeviceId4 := domain.DeviceId("device1")
	remoteDeviceId4 := domain.DeviceId("device2")
	rootName4 := domain.RootName("root")
	err = s.repo.CreateBaseSnapshot(ctx, baseSnapshot4, localDeviceId4, remoteDeviceId4, rootName4)
	s.Require().NoError(err)

	snapshot, err := s.repo.GetLastBaseSnapshotByDeviceIdAndRootName(ctx, localDeviceId1, remoteDeviceId1, rootName1)
	s.Require().NoError(err)
	s.Require().Equal(baseSnapshot4, snapshot)
}

func (s *BaseRepositorySuite) TestGet_ReturnsDbError() {
	db, _ := s.db.DB()
	db.Close()

	_, err := s.repo.GetLastBaseSnapshotByDeviceIdAndRootName(
		context.Background(),
		"device1", "device2", "root",
	)

	s.Require().Error(err)
	s.Require().NotErrorIs(err, baserepo.ErrBaseSnapshotNotFound)
}

func (s *BaseRepositorySuite) TestNewRepository_InvalidOptions() {
	repo, err := baserepo.NewBaseSnapshotRepository(baserepo.BaseSnapshotRepositoryOptions{})
	s.Require().ErrorIs(err, baserepo.ErrInvalidBaseSnapshotRepositoryOptions)
	s.Require().Nil(repo)
}

func (s *BaseRepositorySuite) getDefaultSnapshot() domain.Snapshot {
	s.T().Helper()

	return domain.Snapshot{
		Files: s.getDefaultFileEntries(),
	}
}

func (s *BaseRepositorySuite) getDefaultFileEntries() []domain.FileEntry {
	s.T().Helper()

	path1, err := domain.NewPath("path")
	s.Require().NoError(err)
	path2, err := domain.NewPath("path/to")
	s.Require().NoError(err)
	path3, err := domain.NewPath("path/to/file2")
	s.Require().NoError(err)

	hash1 := "hash1"
	hash2 := "hash2"
	hash3 := "hash3"

	return []domain.FileEntry{
		{
			RelativePath: path1,
			SubtreeSize:  3,
			FileInfo:     s.getDefaultFileInfo(hash1, true),
		},
		{
			RelativePath: path2,
			SubtreeSize:  2,
			FileInfo:     s.getDefaultFileInfo(hash2, true),
		},
		{
			RelativePath: path3,
			SubtreeSize:  1,
			FileInfo:     s.getDefaultFileInfo(hash3, false),
		},
	}
}

func (s *BaseRepositorySuite) getDefaultFileInfo(hash string, isDir bool) domain.FileInfo {
	s.T().Helper()

	return domain.FileInfo{
		Hash:     hash,
		Metadata: s.getMetadata(isDir, 1, 1),
	}
}

func (s *BaseRepositorySuite) getMetadata(isDir bool, size uint64, mod uint64) domain.FileMetadata {
	s.T().Helper()

	return domain.FileMetadata{
		IsDirectory:  isDir,
		SizeBytes:    size,
		ModifiedUnix: mod,
	}
}
