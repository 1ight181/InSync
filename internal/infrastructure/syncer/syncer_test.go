package syncer

import (
	"context"
	"errors"
	"insync/internal/domain"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type SyncerSuite struct {
	suite.Suite

	mockChangeApplier                 *MockIChangeApplier
	mockConflictResolver              *MockIConflictResolver
	mockPostSyncBaseSnapshotPersister *MockIPostSyncBaseSnapshotPersister
	syncer                            *Syncer
}

func (s *SyncerSuite) SetupTest() {
	s.mockChangeApplier = &MockIChangeApplier{}
	s.mockConflictResolver = &MockIConflictResolver{}
	s.mockPostSyncBaseSnapshotPersister = &MockIPostSyncBaseSnapshotPersister{}

	s.syncer = &Syncer{
		changeApplier:                 s.mockChangeApplier,
		conflictResolver:              s.mockConflictResolver,
		postSyncBaseSnapshotPersister: s.mockPostSyncBaseSnapshotPersister,
	}
}

func (s *SyncerSuite) TearDownTest() {
	s.mockChangeApplier.AssertExpectations(s.T())
	s.mockConflictResolver.AssertExpectations(s.T())
	s.mockPostSyncBaseSnapshotPersister.AssertExpectations(s.T())
}

func TestSyncer(t *testing.T) {
	suite.Run(t, new(SyncerSuite))
}

func (s *SyncerSuite) TestNewSyncer_Success() {
	opts := SyncerOptions{
		ChangeApplier:                 s.mockChangeApplier,
		ConflictResolver:              s.mockConflictResolver,
		PostSyncBaseSnapshotPersister: s.mockPostSyncBaseSnapshotPersister,
	}

	syncer, err := NewSyncer(opts)
	s.Require().NoError(err)
	s.NotNil(syncer)
}

func (s *SyncerSuite) TestNewSyncer_NilChangeApplier_ReturnsError() {
	opts := SyncerOptions{
		ConflictResolver:              s.mockConflictResolver,
		PostSyncBaseSnapshotPersister: s.mockPostSyncBaseSnapshotPersister,
	}

	syncer, err := NewSyncer(opts)
	s.Require().ErrorIs(err, ErrInvalidSyncerOptions)
	s.Nil(syncer)
}

func (s *SyncerSuite) TestNewSyncer_NilConflictResolver_ReturnsError() {
	opts := SyncerOptions{
		ChangeApplier:                 s.mockChangeApplier,
		PostSyncBaseSnapshotPersister: s.mockPostSyncBaseSnapshotPersister,
	}

	syncer, err := NewSyncer(opts)
	s.Require().ErrorIs(err, ErrInvalidSyncerOptions)
	s.Nil(syncer)
}

func (s *SyncerSuite) TestNewSyncer_NilPostSyncBaseSnapshotPersister_ReturnsError() {
	opts := SyncerOptions{
		ChangeApplier:    s.mockChangeApplier,
		ConflictResolver: s.mockConflictResolver,
	}

	syncer, err := NewSyncer(opts)
	s.Require().ErrorIs(err, ErrInvalidSyncerOptions)
	s.Nil(syncer)
}

func (s *SyncerSuite) TestSync_AppliesLocalAndRemoteChanges_Success() {
	ctx := context.Background()
	rootName := domain.RootName("test")

	localChange := domain.LocalChange{
		NewRelativePath: mustPath(s.T(), "/file.txt"),
		ChangeType:      domain.CreateFile,
	}
	remoteChange := domain.RemoteChange{
		NewRelativePath: mustPath(s.T(), "/remote.txt"),
		ChangeType:      domain.CreateFile,
	}

	plan := domain.NewSyncPlan([]domain.LocalChange{localChange}, []domain.RemoteChange{remoteChange}, nil)

	s.mockChangeApplier.On("ApplyLocal", ctx, rootName, localChange).Return(nil)
	s.mockChangeApplier.On("ApplyRemote", ctx, rootName, remoteChange).Return(nil)
	s.mockPostSyncBaseSnapshotPersister.On("UpdateBaseSnapshot", ctx, rootName).Return(nil)

	appliedChanges, conflicts, errors, _, err := s.syncer.Sync(ctx, plan, rootName)
	s.Require().NoError(err)

	// Collect results
	var changeEvents []domain.ChangeEvent
	var conflictList []domain.Conflict
	var errorList []error

	done := make(chan struct{})
	go func() {
		for event := range appliedChanges {
			changeEvents = append(changeEvents, event)
		}
		for conflict := range conflicts {
			conflictList = append(conflictList, conflict)
		}
		for err := range errors {
			errorList = append(errorList, err)
		}
		close(done)
	}()

	<-done

	s.Len(changeEvents, 2)
	s.Len(conflictList, 0)
	s.Len(errorList, 0)
}

func (s *SyncerSuite) TestSync_ApplyLocalChange_Fails_SendsError() {
	ctx := context.Background()
	rootName := domain.RootName("test")

	localChange := domain.LocalChange{
		NewRelativePath: mustPath(s.T(), "/file.txt"),
		ChangeType:      domain.CreateFile,
	}

	plan := domain.NewSyncPlan([]domain.LocalChange{localChange}, nil, nil)

	expectedErr := errors.New("apply local error")
	s.mockChangeApplier.On("ApplyLocal", ctx, rootName, localChange).Return(expectedErr)
	s.mockPostSyncBaseSnapshotPersister.On("UpdateBaseSnapshot", ctx, rootName).Return(nil)

	appliedChanges, conflicts, errors, _, err := s.syncer.Sync(ctx, plan, rootName)
	s.Require().NoError(err)

	var changeEvents []domain.ChangeEvent
	done := make(chan struct{})
	go func() {
		for event := range appliedChanges {
			changeEvents = append(changeEvents, event)
		}
		<-conflicts
		<-errors
		close(done)
	}()

	<-done

	s.Len(changeEvents, 1)
	s.ErrorIs(changeEvents[0].Err, expectedErr)
}

func (s *SyncerSuite) TestSync_WithConflicts_ResolvesSuccessfully() {
	ctx := context.Background()
	rootName := domain.RootName("test")

	conflict := domain.Conflict{
		LocalRelativePath:  mustPath(s.T(), "/file.txt"),
		RemoteRelativePath: mustPath(s.T(), "/file.txt"),
		Conflict:           domain.ConflictBothModifiedAtSameTime,
	}

	plan := domain.NewSyncPlan(nil, nil, []domain.Conflict{conflict})

	decision := domain.LocalWin
	requiredChange := domain.SyncChange{
		NewRelativePath: mustPath(s.T(), "/file.txt"),
		ChangeType:      domain.Modify,
	}

	s.mockConflictResolver.On("Resolve", conflict, decision).Return(requiredChange, true, nil)
	s.mockChangeApplier.On("ApplyLocal", ctx, rootName, requiredChange.ToLocalChange()).Return(nil)
	s.mockPostSyncBaseSnapshotPersister.On("UpdateBaseSnapshot", ctx, rootName).Return(nil)

	appliedChanges, conflicts, errors, userDecision, err := s.syncer.Sync(ctx, plan, rootName)
	s.Require().NoError(err)

	// Send decision
	userDecision <- decision
	close(userDecision)

	var changeEvents []domain.ChangeEvent
	var conflictList []domain.Conflict

	done := make(chan struct{})
	go func() {
		for event := range appliedChanges {
			changeEvents = append(changeEvents, event)
		}
		for c := range conflicts {
			conflictList = append(conflictList, c)
		}
		<-errors
		close(done)
	}()

	<-done

	s.Len(conflictList, 1)
	s.Equal(conflict, conflictList[0])
	s.Len(changeEvents, 1)
	s.Equal(requiredChange, changeEvents[0].Change)
}

func (s *SyncerSuite) TestSync_UpdateBaseSnapshot_Fails_SendsError() {
	ctx := context.Background()
	rootName := domain.RootName("test")

	plan := domain.NewSyncPlan(nil, nil, nil)

	expectedErr := errors.New("update snapshot error")
	s.mockPostSyncBaseSnapshotPersister.On("UpdateBaseSnapshot", ctx, rootName).Return(expectedErr)

	appliedChanges, conflicts, errors, _, err := s.syncer.Sync(ctx, plan, rootName)
	s.Require().NoError(err)

	var errorList []error
	done := make(chan struct{})
	go func() {
		<-appliedChanges
		<-conflicts
		for err := range errors {
			errorList = append(errorList, err)
		}
		close(done)
	}()

	<-done

	s.Len(errorList, 1)
	s.ErrorIs(errorList[0], expectedErr)
}

// Mock implementations
type MockIChangeApplier struct {
	mock.Mock
}

func (m *MockIChangeApplier) ApplyLocal(ctx context.Context, rootName domain.RootName, change domain.LocalChange) error {
	args := m.Called(ctx, rootName, change)
	return args.Error(0)
}

func (m *MockIChangeApplier) ApplyRemote(ctx context.Context, rootName domain.RootName, change domain.RemoteChange) error {
	args := m.Called(ctx, rootName, change)
	return args.Error(0)
}

type MockIConflictResolver struct {
	mock.Mock
}

func (m *MockIConflictResolver) Resolve(conflict domain.Conflict, userDecision domain.Decision) (domain.SyncChange, bool, error) {
	args := m.Called(conflict, userDecision)
	return args.Get(0).(domain.SyncChange), args.Bool(1), args.Error(2)
}

type MockIPostSyncBaseSnapshotPersister struct {
	mock.Mock
}

func (m *MockIPostSyncBaseSnapshotPersister) UpdateBaseSnapshot(ctx context.Context, rootName domain.RootName) error {
	args := m.Called(ctx, rootName)
	return args.Error(0)
}
