package syncer

import (
	"insync/internal/domain"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ConflictResolverSuite struct {
	suite.Suite

	resolver *ConflictResolver
}

func (s *ConflictResolverSuite) SetupTest() {
	s.resolver = NewConflictResolver()
}

func TestConflictResolver(t *testing.T) {
	suite.Run(t, new(ConflictResolverSuite))
}

func (s *ConflictResolverSuite) TestResolve_SkipDecision_ReturnsSkipError() {
	conflict := domain.Conflict{
		Conflict: domain.ConflictBothModifiedAtSameTime,
	}
	decision := domain.Skip

	_, _, err := s.resolver.Resolve(conflict, decision)
	s.Require().ErrorIs(err, ErrShouldSkip)
}

func (s *ConflictResolverSuite) TestResolve_ConflictLocalDeletedRemoteModified_LocalWin() {
	conflict := domain.Conflict{
		RemoteRelativePath: mustPath(s.T(), "/file.txt"),
		Conflict:           domain.ConflictLocalDeletedRemoteModified,
	}
	decision := domain.LocalWin

	change, isLocal, err := s.resolver.Resolve(conflict, decision)
	s.Require().NoError(err)
	s.False(isLocal)
	s.Equal(domain.Delete, change.ChangeType)
	s.Equal(conflict.RemoteRelativePath, change.OldRelativePath)
}

func (s *ConflictResolverSuite) TestResolve_ConflictLocalDeletedRemoteModified_RemoteWin() {
	conflict := domain.Conflict{
		RemoteRelativePath: mustPath(s.T(), "/file.txt"),
		Conflict:           domain.ConflictLocalDeletedRemoteModified,
	}
	decision := domain.RemoteWin

	change, isLocal, err := s.resolver.Resolve(conflict, decision)
	s.Require().NoError(err)
	s.True(isLocal)
	s.Equal(domain.CreateFile, change.ChangeType)
	s.Equal(conflict.RemoteRelativePath, change.NewRelativePath)
}

func (s *ConflictResolverSuite) TestResolve_ConflictRemoteDeletedLocalModified_LocalWin() {
	conflict := domain.Conflict{
		LocalRelativePath: mustPath(s.T(), "/file.txt"),
		Conflict:          domain.ConflictRemoteDeletedLocalModified,
	}
	decision := domain.LocalWin

	change, isLocal, err := s.resolver.Resolve(conflict, decision)
	s.Require().NoError(err)
	s.False(isLocal)
	s.Equal(domain.CreateFile, change.ChangeType)
	s.Equal(conflict.LocalRelativePath, change.NewRelativePath)
}

func (s *ConflictResolverSuite) TestResolve_ConflictRemoteDeletedLocalModified_RemoteWin() {
	conflict := domain.Conflict{
		LocalRelativePath: mustPath(s.T(), "/file.txt"),
		Conflict:          domain.ConflictRemoteDeletedLocalModified,
	}
	decision := domain.RemoteWin

	change, isLocal, err := s.resolver.Resolve(conflict, decision)
	s.Require().NoError(err)
	s.True(isLocal)
	s.Equal(domain.Delete, change.ChangeType)
	s.Equal(conflict.LocalRelativePath, change.OldRelativePath)
}

func (s *ConflictResolverSuite) TestResolve_ConflictLocalMovedRemoteMoved_LocalWin() {
	conflict := domain.Conflict{
		LocalRelativePath:  mustPath(s.T(), "/local.txt"),
		RemoteRelativePath: mustPath(s.T(), "/remote.txt"),
		Conflict:           domain.ConflictLocalMovedRemoteMoved,
	}
	decision := domain.LocalWin

	change, isLocal, err := s.resolver.Resolve(conflict, decision)
	s.Require().NoError(err)
	s.False(isLocal)
	s.Equal(domain.Move, change.ChangeType)
	s.Equal(conflict.RemoteRelativePath, change.OldRelativePath)
	s.Equal(conflict.LocalRelativePath, change.NewRelativePath)
}

func (s *ConflictResolverSuite) TestResolve_ConflictLocalRenamedRemoteRenamed_RemoteWin() {
	conflict := domain.Conflict{
		LocalRelativePath:  mustPath(s.T(), "/local.txt"),
		RemoteRelativePath: mustPath(s.T(), "/remote.txt"),
		Conflict:           domain.ConflictLocalRenamedRemoteRenamed,
	}
	decision := domain.RemoteWin

	change, isLocal, err := s.resolver.Resolve(conflict, decision)
	s.Require().NoError(err)
	s.True(isLocal)
	s.Equal(domain.Rename, change.ChangeType)
	s.Equal(conflict.LocalRelativePath, change.OldRelativePath)
	s.Equal(conflict.RemoteRelativePath, change.NewRelativePath)
}

func (s *ConflictResolverSuite) TestResolve_ConflictBothModifiedAtSameTime_LocalWin() {
	conflict := domain.Conflict{
		LocalRelativePath:  mustPath(s.T(), "/file.txt"),
		RemoteRelativePath: mustPath(s.T(), "/file.txt"),
		Conflict:           domain.ConflictBothModifiedAtSameTime,
	}
	decision := domain.LocalWin

	change, isLocal, err := s.resolver.Resolve(conflict, decision)
	s.Require().NoError(err)
	s.False(isLocal)
	s.Equal(domain.Modify, change.ChangeType)
	s.Equal(conflict.RemoteRelativePath, change.OldRelativePath)
}

func (s *ConflictResolverSuite) TestResolve_ConflictBothCreatedAtSamePathConflict_RemoteWin() {
	conflict := domain.Conflict{
		LocalRelativePath: mustPath(s.T(), "/file.txt"),
		Conflict:          domain.ConflictBothCreatedAtSamePathConflict,
	}
	decision := domain.RemoteWin

	change, isLocal, err := s.resolver.Resolve(conflict, decision)
	s.Require().NoError(err)
	s.True(isLocal)
	s.Equal(domain.Modify, change.ChangeType)
	s.Equal(conflict.LocalRelativePath, change.OldRelativePath)
}

func (s *ConflictResolverSuite) TestResolve_UnknownConflictType_ReturnsError() {
	conflict := domain.Conflict{
		Conflict: domain.ConflictType(999), // invalid
	}
	decision := domain.LocalWin

	_, _, err := s.resolver.Resolve(conflict, decision)
	s.Require().ErrorIs(err, ErrUnknownConflictType)
}

func mustPath(t *testing.T, rawPath string) domain.Path {
	t.Helper()
	p, err := domain.NewPath(rawPath)
	require.NoError(t, err)
	return p
}
