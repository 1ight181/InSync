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

	mockBaseSnapshotManager   *MockIBaseSnapshotManager
	mockLocalSnapshotProvider *MockILocalSnapshotProvider
	mockClientFactory         *MockIClientFactory
	persister                 *PostSyncBaseSnapshotPersister
}

func (s *PostSyncBaseSnapshotPersisterSuite) SetupTest() {
	s.mockBaseSnapshotManager = &MockIBaseSnapshotManager{}
	s.mockLocalSnapshotProvider = &MockILocalSnapshotProvider{}
	s.mockClientFactory = &MockIClientFactory{}

	s.persister = &PostSyncBaseSnapshotPersister{
		baseSnapshotManager:   s.mockBaseSnapshotManager,
		localSnapshotProvider: s.mockLocalSnapshotProvider,
		clientFactory:         s.mockClientFactory,
	}
}

func (s *PostSyncBaseSnapshotPersisterSuite) TearDownTest() {
	s.mockBaseSnapshotManager.AssertExpectations(s.T())
	s.mockLocalSnapshotProvider.AssertExpectations(s.T())
	s.mockClientFactory.AssertExpectations(s.T())
}

func TestPostSyncBaseSnapshotPersister(t *testing.T) {
	suite.Run(t, new(PostSyncBaseSnapshotPersisterSuite))
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestNewPostSyncBaseSnapshotPersister_Success() {
	opts := PostSyncBaseSnapshotPersisterOptions{
		BaseSnapshotManager:   s.mockBaseSnapshotManager,
		LocalSnapshotProvider: s.mockLocalSnapshotProvider,
		ClientFactory:         s.mockClientFactory,
	}

	persister, err := NewPostSyncBaseSnapshotPersister(opts)
	s.Require().NoError(err)
	s.NotNil(persister)
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestNewPostSyncBaseSnapshotPersister_NilBaseSnapshotRepository_ReManager() {
	opts := PostSyncBaseSnapshotPersisterOptions{
		LocalSnapshotProvider: s.mockLocalSnapshotProvider,
		ClientFactory:         s.mockClientFactory,
	}

	persister, err := NewPostSyncBaseSnapshotPersister(opts)
	s.Require().ErrorIs(err, ErrInvalidOpts)
	s.Nil(persister)
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestNewPostSyncBaseSnapshotPersister_NilSnapshotProvider_ReturnsError() {
	opts := PostSyncBaseSnapshotPersisterOptions{
		BaseSnapshotManager: s.mockBaseSnapshotManager,
		ClientFactory:       s.mockClientFactory,
	}

	persister, err := NewPostSyncBaseSnapshotPersister(opts)
	s.Require().ErrorIs(err, ErrInvalidOpts)
	s.Nil(persister)
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestUpdateBaseSnapshot_Success() {
	ctx := context.Background()
	rootName := domain.RootName("test")
	snapshot := domain.Snapshot{Files: []domain.FileEntry{}}
	baseSnapshot := domain.BaseSnapshot{Snapshot: snapshot, IsInitial: false}

	s.mockBaseSnapshotManager.On("GetBaseSnapshot", ctx, rootName).Return(baseSnapshot, nil)
	s.mockBaseSnapshotManager.On("CreateBaseSnapshot", ctx, snapshot, rootName, false).Return(nil)
	s.mockLocalSnapshotProvider.On("GetLocalSnapshot", ctx, rootName, mock.Anything).Return(snapshot, nil)
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
	snapshot := domain.Snapshot{Files: []domain.FileEntry{}}
	baseSnapshot := domain.BaseSnapshot{Snapshot: snapshot, IsInitial: false}

	s.mockBaseSnapshotManager.On("GetBaseSnapshot", ctx, rootName).Return(baseSnapshot, nil)
	s.mockLocalSnapshotProvider.On("GetLocalSnapshot", ctx, rootName, mock.Anything).Return(domain.Snapshot{}, expectedErr)

	err := s.persister.UpdateBaseSnapshot(ctx, rootName)
	s.Require().ErrorIs(err, expectedErr)
}
func (s *PostSyncBaseSnapshotPersisterSuite) TestUpdateBaseSnapshot_CreateBaseSnapshot_Fails_ReturnsError() {
	ctx := context.Background()
	rootName := domain.RootName("test")
	snapshot := domain.Snapshot{}
	baseSnapshot := domain.BaseSnapshot{Snapshot: snapshot, IsInitial: false}

	s.mockBaseSnapshotManager.On("GetBaseSnapshot", ctx, rootName).Return(baseSnapshot, nil)
	s.mockLocalSnapshotProvider.On("GetLocalSnapshot", ctx, rootName, mock.Anything).Return(snapshot, nil)
	expectedErr := errors.New("create snapshot error")
	s.mockBaseSnapshotManager.On("CreateBaseSnapshot", ctx, snapshot, rootName, false).Return(expectedErr)

	err := s.persister.UpdateBaseSnapshot(ctx, rootName)
	s.Require().ErrorIs(err, expectedErr)
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestUpdateBaseSnapshot_CurrentClient_Fails_ReturnsError() {
	ctx := context.Background()
	rootName := domain.RootName("test")
	snapshot := domain.Snapshot{}
	baseSnapshot := domain.BaseSnapshot{Snapshot: snapshot, IsInitial: false}

	s.mockBaseSnapshotManager.On("GetBaseSnapshot", ctx, rootName).Return(baseSnapshot, nil)
	s.mockLocalSnapshotProvider.On("GetLocalSnapshot", ctx, rootName, mock.Anything).Return(snapshot, nil)
	s.mockBaseSnapshotManager.On("CreateBaseSnapshot", ctx, snapshot, rootName, false).Return(nil)

	expectedErr := errors.New("current client error")
	s.mockClientFactory.On("CurrentClient").Return(interfaces.IClient(nil), expectedErr)

	err := s.persister.UpdateBaseSnapshot(ctx, rootName)
	s.Require().ErrorIs(err, expectedErr)
}

func (s *PostSyncBaseSnapshotPersisterSuite) TestUpdateBaseSnapshot_RemoteUpdate_Fails_ReturnsError() {
	ctx := context.Background()
	rootName := domain.RootName("test")
	snapshot := domain.Snapshot{}
	baseSnapshot := domain.BaseSnapshot{Snapshot: snapshot, IsInitial: false}

	s.mockBaseSnapshotManager.On("GetBaseSnapshot", ctx, rootName).Return(baseSnapshot, nil)
	s.mockLocalSnapshotProvider.On("GetLocalSnapshot", ctx, rootName, mock.Anything).Return(snapshot, nil)
	s.mockBaseSnapshotManager.On("CreateBaseSnapshot", ctx, snapshot, rootName, false).Return(nil)

	mockClient := &MockIClient{}
	expectedErr := errors.New("remote update error")
	mockClient.On("UpdateBaseSnapshot", ctx, rootName).Return(expectedErr)
	s.mockClientFactory.On("CurrentClient").Return(mockClient, nil)

	err := s.persister.UpdateBaseSnapshot(ctx, rootName)
	s.Require().ErrorIs(err, expectedErr)
}

// Mock implementations
type MockIBaseSnapshotManager struct {
	mock.Mock
}

func (m *MockIBaseSnapshotManager) CreateBaseSnapshot(ctx context.Context, snapshot domain.Snapshot, rootName domain.RootName, isInitial bool) error {
	args := m.Called(ctx, snapshot, rootName, isInitial)
	return args.Error(0)
}

func (m *MockIBaseSnapshotManager) GetBaseSnapshot(ctx context.Context, rootName domain.RootName) (domain.BaseSnapshot, error) {
	args := m.Called(ctx, rootName)
	return args.Get(0).(domain.BaseSnapshot), args.Error(1)
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

func (m *MockILocalSnapshotProvider) GetLocalSnapshot(ctx context.Context, rootName domain.RootName, baseSnapshot *domain.BaseSnapshot) (domain.Snapshot, error) {
	args := m.Called(ctx, rootName, baseSnapshot)
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

func (m *MockIBaseSnapshotProvider) GetBaseSnapshot(ctx context.Context, rootName domain.RootName) (domain.BaseSnapshot, error) {
	args := m.Called(ctx, rootName)
	return args.Get(0).(domain.BaseSnapshot), args.Error(1)
}
