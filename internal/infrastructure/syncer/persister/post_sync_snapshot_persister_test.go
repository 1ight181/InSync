package base

import (
	"context"
	"errors"
	"insync/internal/domain"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type PostSyncBaseSnapshotPersisterSuite struct {
	suite.Suite

	mockBaseSnapshotRepository *MockIBaseSnapshotRepositoryWriter
	mockDeviceIdProvider       *MockIDeviceIdProvider
	mockSnapshotProvider       *MockIBaseSnapshotProvider
	persister                  *PostSyncBaseSnapshotPersister
}

func (s *PostSyncBaseSnapshotPersisterSuite) SetupTest() {
	s.mockBaseSnapshotRepository = &MockIBaseSnapshotRepositoryWriter{}
	s.mockDeviceIdProvider = &MockIDeviceIdProvider{}
	s.mockSnapshotProvider = &MockIBaseSnapshotProvider{}

	s.persister = &PostSyncBaseSnapshotPersister{
		baseSnapshotRepository: s.mockBaseSnapshotRepository,
		deviceIdProvider:       s.mockDeviceIdProvider,
		snapshotProvider:       s.mockSnapshotProvider,
	}
}

func (s *PostSyncBaseSnapshotPersisterSuite) TearDownTest() {
	s.mockBaseSnapshotRepository.AssertExpectations(s.T())
	s.mockDeviceIdProvider.AssertExpectations(s.T())
	s.mockSnapshotProvider.AssertExpectations(s.T())
}

func TestPostSyncBaseSnapshotPersister(t *testing.T) {
	suite.Run(t, new(PostSyncBaseSnapshotPersisterSuite))
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestNewPostSyncBaseSnapshotPersister_Success() {
	opts := PostSyncBaseSnapshotPersisterOptions{
		BaseSnapshotRepository: s.mockBaseSnapshotRepository,
		DeviceIdProvider:       s.mockDeviceIdProvider,
		SnapshotProvider:       s.mockSnapshotProvider,
	}

	persister, err := NewPostSyncBaseSnapshotPersister(opts)
	s.Require().NoError(err)
	s.NotNil(persister)
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestNewPostSyncBaseSnapshotPersister_NilBaseSnapshotRepository_ReturnsError() {
	opts := PostSyncBaseSnapshotPersisterOptions{
		DeviceIdProvider: s.mockDeviceIdProvider,
		SnapshotProvider: s.mockSnapshotProvider,
	}

	persister, err := NewPostSyncBaseSnapshotPersister(opts)
	s.Require().ErrorIs(err, ErrInvalidOpts)
	s.Nil(persister)
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestNewPostSyncBaseSnapshotPersister_NilDeviceIdProvider_ReturnsError() {
	opts := PostSyncBaseSnapshotPersisterOptions{
		BaseSnapshotRepository: s.mockBaseSnapshotRepository,
		SnapshotProvider:       s.mockSnapshotProvider,
	}

	persister, err := NewPostSyncBaseSnapshotPersister(opts)
	s.Require().ErrorIs(err, ErrInvalidOpts)
	s.Nil(persister)
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestNewPostSyncBaseSnapshotPersister_NilSnapshotProvider_ReturnsError() {
	opts := PostSyncBaseSnapshotPersisterOptions{
		BaseSnapshotRepository: s.mockBaseSnapshotRepository,
		DeviceIdProvider:       s.mockDeviceIdProvider,
	}

	persister, err := NewPostSyncBaseSnapshotPersister(opts)
	s.Require().ErrorIs(err, ErrInvalidOpts)
	s.Nil(persister)
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestUpdateBaseSnapshot_Success() {
	ctx := context.Background()
	rootName := domain.RootName("test")
	snapshot := domain.Snapshot{Files: []domain.FileEntry{}}
	remoteDeviceId := domain.DeviceId("remote-id")
	localDeviceId := domain.DeviceId("local-id")

	s.mockSnapshotProvider.On("GetBaseSnapshot", ctx, rootName).Return(snapshot, nil)
	s.mockDeviceIdProvider.On("GetCurrentRemoteDeviceId").Return(remoteDeviceId, nil)
	s.mockDeviceIdProvider.On("GetCurrentLocalDeviceId").Return(localDeviceId, nil)
	s.mockBaseSnapshotRepository.On("CreateBaseSnapshot", ctx, snapshot, localDeviceId, remoteDeviceId, rootName).Return(nil)

	err := s.persister.UpdateBaseSnapshot(ctx, rootName)
	s.Require().NoError(err)
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestUpdateBaseSnapshot_GetBaseSnapshot_Fails_ReturnsError() {
	ctx := context.Background()
	rootName := domain.RootName("test")

	expectedErr := errors.New("get snapshot error")
	s.mockSnapshotProvider.On("GetBaseSnapshot", ctx, rootName).Return(domain.Snapshot{}, expectedErr)

	err := s.persister.UpdateBaseSnapshot(ctx, rootName)
	s.Require().ErrorIs(err, expectedErr)
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestUpdateBaseSnapshot_GetRemoteDeviceId_Fails_ReturnsError() {
	ctx := context.Background()
	rootName := domain.RootName("test")
	snapshot := domain.Snapshot{}

	s.mockSnapshotProvider.On("GetBaseSnapshot", ctx, rootName).Return(snapshot, nil)
	expectedErr := errors.New("get remote id error")
	s.mockDeviceIdProvider.On("GetCurrentRemoteDeviceId").Return(domain.DeviceId(""), expectedErr)

	err := s.persister.UpdateBaseSnapshot(ctx, rootName)
	s.Require().ErrorIs(err, expectedErr)
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestUpdateBaseSnapshot_GetLocalDeviceId_Fails_ReturnsError() {
	ctx := context.Background()
	rootName := domain.RootName("test")
	snapshot := domain.Snapshot{}
	remoteDeviceId := domain.DeviceId("remote-id")

	s.mockSnapshotProvider.On("GetBaseSnapshot", ctx, rootName).Return(snapshot, nil)
	s.mockDeviceIdProvider.On("GetCurrentRemoteDeviceId").Return(remoteDeviceId, nil)
	expectedErr := errors.New("get local id error")
	s.mockDeviceIdProvider.On("GetCurrentLocalDeviceId").Return(domain.DeviceId(""), expectedErr)

	err := s.persister.UpdateBaseSnapshot(ctx, rootName)
	s.Require().ErrorIs(err, expectedErr)
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestUpdateBaseSnapshot_CreateBaseSnapshot_Fails_ReturnsError() {
	ctx := context.Background()
	rootName := domain.RootName("test")
	snapshot := domain.Snapshot{}
	remoteDeviceId := domain.DeviceId("remote-id")
	localDeviceId := domain.DeviceId("local-id")

	s.mockSnapshotProvider.On("GetBaseSnapshot", ctx, rootName).Return(snapshot, nil)
	s.mockDeviceIdProvider.On("GetCurrentRemoteDeviceId").Return(remoteDeviceId, nil)
	s.mockDeviceIdProvider.On("GetCurrentLocalDeviceId").Return(localDeviceId, nil)
	expectedErr := errors.New("create snapshot error")
	s.mockBaseSnapshotRepository.On("CreateBaseSnapshot", ctx, snapshot, localDeviceId, remoteDeviceId, rootName).Return(expectedErr)

	err := s.persister.UpdateBaseSnapshot(ctx, rootName)
	s.Require().ErrorIs(err, expectedErr)
}

// Mock implementations
type MockIBaseSnapshotRepositoryWriter struct {
	mock.Mock
}

func (m *MockIBaseSnapshotRepositoryWriter) CreateBaseSnapshot(ctx context.Context, snapshot domain.Snapshot, localDeviceId domain.DeviceId, remoteDeviceId domain.DeviceId, rootName domain.RootName) error {
	args := m.Called(ctx, snapshot, localDeviceId, remoteDeviceId, rootName)
	return args.Error(0)
}

type MockIDeviceIdProvider struct {
	mock.Mock
}

func (m *MockIDeviceIdProvider) GetCurrentRemoteDeviceId() (domain.DeviceId, error) {
	args := m.Called()
	return args.Get(0).(domain.DeviceId), args.Error(1)
}

func (m *MockIDeviceIdProvider) GetCurrentLocalDeviceId() (domain.DeviceId, error) {
	args := m.Called()
	return args.Get(0).(domain.DeviceId), args.Error(1)
}

type MockIBaseSnapshotProvider struct {
	mock.Mock
}

func (m *MockIBaseSnapshotProvider) GetBaseSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error) {
	args := m.Called(ctx, rootName)
	return args.Get(0).(domain.Snapshot), args.Error(1)
}
