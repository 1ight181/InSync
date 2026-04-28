package base

import (
	"context"
	"errors"
	"insync/internal/domain"
	"insync/internal/interfaces"
	"io"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type PostSyncBaseSnapshotPersisterSuite struct {
	suite.Suite

	mockBaseSnapshotRepository *MockIBaseSnapshotRepositoryWriter
	mockDeviceIdProvider       *MockIDeviceIdProvider
	mockLocalSnapshotProvider  *MockILocalSnapshotProvider
	mockClientFactory          *MockIClientFactory
	persister                  *PostSyncBaseSnapshotPersister
}

func (s *PostSyncBaseSnapshotPersisterSuite) SetupTest() {
	s.mockBaseSnapshotRepository = &MockIBaseSnapshotRepositoryWriter{}
	s.mockDeviceIdProvider = &MockIDeviceIdProvider{}
	s.mockLocalSnapshotProvider = &MockILocalSnapshotProvider{}
	s.mockClientFactory = &MockIClientFactory{}

	s.persister = &PostSyncBaseSnapshotPersister{
		baseSnapshotRepository: s.mockBaseSnapshotRepository,
		deviceIdProvider:       s.mockDeviceIdProvider,
		localSnapshotProvider:  s.mockLocalSnapshotProvider,
		clientFactory:          s.mockClientFactory,
	}
}

func (s *PostSyncBaseSnapshotPersisterSuite) TearDownTest() {
	s.mockBaseSnapshotRepository.AssertExpectations(s.T())
	s.mockDeviceIdProvider.AssertExpectations(s.T())
	s.mockLocalSnapshotProvider.AssertExpectations(s.T())
	s.mockClientFactory.AssertExpectations(s.T())
}

func TestPostSyncBaseSnapshotPersister(t *testing.T) {
	suite.Run(t, new(PostSyncBaseSnapshotPersisterSuite))
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestNewPostSyncBaseSnapshotPersister_Success() {
	opts := PostSyncBaseSnapshotPersisterOptions{
		BaseSnapshotRepository: s.mockBaseSnapshotRepository,
		DeviceIdProvider:       s.mockDeviceIdProvider,
		LocalSnapshotProvider:  s.mockLocalSnapshotProvider,
		ClientFactory:          s.mockClientFactory,
	}

	persister, err := NewPostSyncBaseSnapshotPersister(opts)
	s.Require().NoError(err)
	s.NotNil(persister)
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestNewPostSyncBaseSnapshotPersister_NilBaseSnapshotRepository_ReturnsError() {
	opts := PostSyncBaseSnapshotPersisterOptions{
		DeviceIdProvider:      s.mockDeviceIdProvider,
		LocalSnapshotProvider: s.mockLocalSnapshotProvider,
		ClientFactory:         s.mockClientFactory,
	}

	persister, err := NewPostSyncBaseSnapshotPersister(opts)
	s.Require().ErrorIs(err, ErrInvalidOpts)
	s.Nil(persister)
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestNewPostSyncBaseSnapshotPersister_NilDeviceIdProvider_ReturnsError() {
	opts := PostSyncBaseSnapshotPersisterOptions{
		BaseSnapshotRepository: s.mockBaseSnapshotRepository,
		LocalSnapshotProvider:  s.mockLocalSnapshotProvider,
		ClientFactory:          s.mockClientFactory,
	}

	persister, err := NewPostSyncBaseSnapshotPersister(opts)
	s.Require().ErrorIs(err, ErrInvalidOpts)
	s.Nil(persister)
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestNewPostSyncBaseSnapshotPersister_NilSnapshotProvider_ReturnsError() {
	opts := PostSyncBaseSnapshotPersisterOptions{
		BaseSnapshotRepository: s.mockBaseSnapshotRepository,
		DeviceIdProvider:       s.mockDeviceIdProvider,
		ClientFactory:          s.mockClientFactory,
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

	s.mockLocalSnapshotProvider.On("GetLocalSnapshot", ctx, rootName).Return(snapshot, nil)
	s.mockDeviceIdProvider.On("GetCurrentRemoteDeviceId").Return(remoteDeviceId, nil)
	s.mockDeviceIdProvider.On("GetCurrentLocalDeviceId").Return(localDeviceId, nil)
	s.mockBaseSnapshotRepository.On("CreateBaseSnapshot", ctx, snapshot, localDeviceId, remoteDeviceId, rootName).Return(nil)
	mockClient := &MockIClient{}
	mockClient.On("UpdateBaseSnapshot", ctx, rootName).Return(nil)
	s.mockClientFactory.On("CurrentClient").Return(mockClient, nil)

	err := s.persister.UpdateBaseSnapshot(ctx, rootName)
	s.Require().NoError(err)
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestUpdateBaseSnapshot_GetBaseSnapshot_Fails_ReturnsError() {
	ctx := context.Background()
	rootName := domain.RootName("test")

	expectedErr := errors.New("get snapshot error")
	s.mockLocalSnapshotProvider.On("GetLocalSnapshot", ctx, rootName).Return(domain.Snapshot{}, expectedErr)

	err := s.persister.UpdateBaseSnapshot(ctx, rootName)
	s.Require().ErrorIs(err, expectedErr)
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestUpdateBaseSnapshot_GetRemoteDeviceId_Fails_ReturnsError() {
	ctx := context.Background()
	rootName := domain.RootName("test")
	snapshot := domain.Snapshot{}

	s.mockLocalSnapshotProvider.On("GetLocalSnapshot", ctx, rootName).Return(snapshot, nil)
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

	s.mockLocalSnapshotProvider.On("GetLocalSnapshot", ctx, rootName).Return(snapshot, nil)
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

	s.mockLocalSnapshotProvider.On("GetLocalSnapshot", ctx, rootName).Return(snapshot, nil)
	s.mockDeviceIdProvider.On("GetCurrentRemoteDeviceId").Return(remoteDeviceId, nil)
	s.mockDeviceIdProvider.On("GetCurrentLocalDeviceId").Return(localDeviceId, nil)
	expectedErr := errors.New("create snapshot error")
	s.mockBaseSnapshotRepository.On("CreateBaseSnapshot", ctx, snapshot, localDeviceId, remoteDeviceId, rootName).Return(expectedErr)

	err := s.persister.UpdateBaseSnapshot(ctx, rootName)
	s.Require().ErrorIs(err, expectedErr)
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestUpdateBaseSnapshot_CurrentClient_Fails_ReturnsError() {
	ctx := context.Background()
	rootName := domain.RootName("test")
	snapshot := domain.Snapshot{}
	remoteDeviceId := domain.DeviceId("remote-id")
	localDeviceId := domain.DeviceId("local-id")

	s.mockLocalSnapshotProvider.On("GetLocalSnapshot", ctx, rootName).Return(snapshot, nil)
	s.mockDeviceIdProvider.On("GetCurrentRemoteDeviceId").Return(remoteDeviceId, nil)
	s.mockDeviceIdProvider.On("GetCurrentLocalDeviceId").Return(localDeviceId, nil)
	s.mockBaseSnapshotRepository.On("CreateBaseSnapshot", ctx, snapshot, localDeviceId, remoteDeviceId, rootName).Return(nil)

	expectedErr := errors.New("current client error")
	s.mockClientFactory.On("CurrentClient").Return(interfaces.IClient(nil), expectedErr)

	err := s.persister.UpdateBaseSnapshot(ctx, rootName)
	s.Require().ErrorIs(err, expectedErr)
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestUpdateBaseSnapshot_RemoteUpdate_Fails_ReturnsError() {
	ctx := context.Background()
	rootName := domain.RootName("test")
	snapshot := domain.Snapshot{}
	remoteDeviceId := domain.DeviceId("remote-id")
	localDeviceId := domain.DeviceId("local-id")

	s.mockLocalSnapshotProvider.On("GetLocalSnapshot", ctx, rootName).Return(snapshot, nil)
	s.mockDeviceIdProvider.On("GetCurrentRemoteDeviceId").Return(remoteDeviceId, nil)
	s.mockDeviceIdProvider.On("GetCurrentLocalDeviceId").Return(localDeviceId, nil)
	s.mockBaseSnapshotRepository.On("CreateBaseSnapshot", ctx, snapshot, localDeviceId, remoteDeviceId, rootName).Return(nil)

	mockClient := &MockIClient{}
	expectedErr := errors.New("remote update error")
	mockClient.On("UpdateBaseSnapshot", ctx, rootName).Return(expectedErr)
	s.mockClientFactory.On("CurrentClient").Return(mockClient, nil)

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

type MockILocalSnapshotProvider struct {
	mock.Mock
}

func (m *MockILocalSnapshotProvider) GetLocalSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error) {
	args := m.Called(ctx, rootName)
	return args.Get(0).(domain.Snapshot), args.Error(1)
}

type MockIClientFactory struct {
	mock.Mock
}

func (m *MockIClientFactory) CurrentClient() (interfaces.IClient, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(interfaces.IClient), args.Error(1)
}

type MockIClient struct {
	mock.Mock
}

func (m *MockIClient) GetSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error) {
	args := m.Called(ctx, rootName)
	return args.Get(0).(domain.Snapshot), args.Error(1)
}

func (m *MockIClient) GetFile(ctx context.Context, scopedPath domain.ScopedPath) (io.ReadCloser, error) {
	args := m.Called(ctx, scopedPath)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockIClient) PutFile(ctx context.Context, file io.Reader, scopedPath domain.ScopedPath) error {
	args := m.Called(ctx, file, scopedPath)
	return args.Error(0)
}

func (m *MockIClient) DeleteFile(ctx context.Context, scopedPath domain.ScopedPath) error {
	args := m.Called(ctx, scopedPath)
	return args.Error(0)
}

func (m *MockIClient) RenameFile(ctx context.Context, oldScopedPath domain.ScopedPath, newScopedPath domain.ScopedPath) error {
	args := m.Called(ctx, oldScopedPath, newScopedPath)
	return args.Error(0)
}

func (m *MockIClient) CreateDir(ctx context.Context, scopedPath domain.ScopedPath) error {
	args := m.Called(ctx, scopedPath)
	return args.Error(0)
}

func (m *MockIClient) UpdateBaseSnapshot(ctx context.Context, rootName domain.RootName) error {
	args := m.Called(ctx, rootName)
	return args.Error(0)
}

type MockIBaseSnapshotProvider struct {
	mock.Mock
}

func (m *MockIBaseSnapshotProvider) GetBaseSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error) {
	args := m.Called(ctx, rootName)
	return args.Get(0).(domain.Snapshot), args.Error(1)
}
