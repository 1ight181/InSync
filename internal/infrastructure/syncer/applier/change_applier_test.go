package local

import (
	"bytes"
	"context"
	"insync/internal/domain"
	"insync/internal/interfaces"
	"io"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ChangeApplierSuite struct {
	suite.Suite

	mockFileManager   *MockIFileManager
	mockClientFactory *MockIClientFactory
	mockClient        *MockIClient
	changeApplier     *ChangeApplier
}

func (s *ChangeApplierSuite) SetupTest() {
	s.mockFileManager = &MockIFileManager{}
	s.mockClientFactory = &MockIClientFactory{}
	s.mockClient = &MockIClient{}

	s.changeApplier = &ChangeApplier{
		fileManager:   s.mockFileManager,
		clientFactory: s.mockClientFactory,
	}
}

func (s *ChangeApplierSuite) TearDownTest() {
	s.mockFileManager.AssertExpectations(s.T())
	s.mockClientFactory.AssertExpectations(s.T())
	s.mockClient.AssertExpectations(s.T())
}

func TestChangeApplier(t *testing.T) {
	suite.Run(t, new(ChangeApplierSuite))
}

func (s *ChangeApplierSuite) TestNewChangeApplier_Success() {
	opts := ChangeApplierOptions{
		FileManager:   s.mockFileManager,
		ClientFactory: s.mockClientFactory,
	}

	applier, err := NewChangeApplier(opts)
	s.Require().NoError(err)
	s.NotNil(applier)
}

func (s *ChangeApplierSuite) TestNewChangeApplier_NilFileManager_ReturnsError() {
	opts := ChangeApplierOptions{
		ClientFactory: s.mockClientFactory,
	}

	applier, err := NewChangeApplier(opts)
	s.Require().ErrorIs(err, ErrInvalidOpts)
	s.Nil(applier)
}

func (s *ChangeApplierSuite) TestNewChangeApplier_NilClientFactory_ReturnsError() {
	opts := ChangeApplierOptions{
		FileManager: s.mockFileManager,
	}

	applier, err := NewChangeApplier(opts)
	s.Require().ErrorIs(err, ErrInvalidOpts)
	s.Nil(applier)
}

func (s *ChangeApplierSuite) TestApplyLocal_CreateFile_Success() {
	ctx := context.Background()
	rootName := domain.RootName("test")
	change := domain.LocalChange{
		NewRelativePath: mustPath(s.T(), "/file.txt"),
		ChangeType:      domain.CreateFile,
	}
	scopedPath := mustScopedPath(s.T(), rootName, change.NewRelativePath)

	content := io.NopCloser(bytes.NewReader([]byte("content")))

	s.mockClientFactory.On("CurrentClient").Return(s.mockClient, nil)
	s.mockClient.On("GetFile", ctx, scopedPath).Return(content, nil)
	s.mockFileManager.On("PutFile", ctx, scopedPath, mock.MatchedBy(func(r io.Reader) bool { return true })).Return(nil)

	err := s.changeApplier.ApplyLocal(ctx, rootName, change)
	s.Require().NoError(err)
}

func (s *ChangeApplierSuite) TestApplyLocal_Delete_Success() {
	ctx := context.Background()
	rootName := domain.RootName("test")
	change := domain.LocalChange{
		OldRelativePath: mustPath(s.T(), "/file.txt"),
		ChangeType:      domain.Delete,
	}
	scopedPath := mustScopedPath(s.T(), rootName, change.OldRelativePath)

	s.mockFileManager.On("DeleteFile", ctx, scopedPath).Return(nil)

	err := s.changeApplier.ApplyLocal(ctx, rootName, change)
	s.Require().NoError(err)
}

func (s *ChangeApplierSuite) TestApplyLocal_Rename_Success() {
	ctx := context.Background()
	rootName := domain.RootName("test")
	change := domain.LocalChange{
		OldRelativePath: mustPath(s.T(), "/old.txt"),
		NewRelativePath: mustPath(s.T(), "/new.txt"),
		ChangeType:      domain.Rename,
	}
	oldScopedPath := mustScopedPath(s.T(), rootName, change.OldRelativePath)
	newScopedPath := mustScopedPath(s.T(), rootName, change.NewRelativePath)

	s.mockFileManager.On("RenameFile", ctx, oldScopedPath, newScopedPath).Return(nil)

	err := s.changeApplier.ApplyLocal(ctx, rootName, change)
	s.Require().NoError(err)
}

func (s *ChangeApplierSuite) TestApplyLocal_UnknownChangeType_ReturnsError() {
	ctx := context.Background()
	rootName := domain.RootName("test")
	change := domain.LocalChange{
		ChangeType: domain.SyncChangeType(999), // invalid
	}

	err := s.changeApplier.ApplyLocal(ctx, rootName, change)
	s.Require().ErrorIs(err, ErrUnknownChangeType)
}

func (s *ChangeApplierSuite) TestApplyRemote_CreateFile_Success() {
	ctx := context.Background()
	rootName := domain.RootName("test")
	change := domain.RemoteChange{
		NewRelativePath: mustPath(s.T(), "/file.txt"),
		ChangeType:      domain.CreateFile,
	}
	scopedPath := mustScopedPath(s.T(), rootName, change.NewRelativePath)

	content := io.NopCloser(bytes.NewReader([]byte("content")))

	s.mockClientFactory.On("CurrentClient").Return(s.mockClient, nil)
	s.mockFileManager.On("GetFile", ctx, scopedPath).Return(content, nil)
	s.mockClient.On("PutFile", ctx, content, scopedPath).Return(nil)

	err := s.changeApplier.ApplyRemote(ctx, rootName, change)
	s.Require().NoError(err)
}

func (s *ChangeApplierSuite) TestApplyRemote_Delete_Success() {
	ctx := context.Background()
	rootName := domain.RootName("test")
	change := domain.RemoteChange{
		OldRelativePath: mustPath(s.T(), "/file.txt"),
		ChangeType:      domain.Delete,
	}
	scopedPath := mustScopedPath(s.T(), rootName, change.OldRelativePath)

	s.mockClientFactory.On("CurrentClient").Return(s.mockClient, nil)
	s.mockClient.On("DeleteFile", ctx, scopedPath).Return(nil)

	err := s.changeApplier.ApplyRemote(ctx, rootName, change)
	s.Require().NoError(err)
}

func (s *ChangeApplierSuite) TestApplyRemote_Rename_Success() {
	ctx := context.Background()
	rootName := domain.RootName("test")
	change := domain.RemoteChange{
		OldRelativePath: mustPath(s.T(), "/old.txt"),
		NewRelativePath: mustPath(s.T(), "/new.txt"),
		ChangeType:      domain.Rename,
	}
	oldScopedPath := mustScopedPath(s.T(), rootName, change.OldRelativePath)
	newScopedPath := mustScopedPath(s.T(), rootName, change.NewRelativePath)

	s.mockClientFactory.On("CurrentClient").Return(s.mockClient, nil)
	s.mockClient.On("RenameFile", ctx, oldScopedPath, newScopedPath).Return(nil)

	err := s.changeApplier.ApplyRemote(ctx, rootName, change)
	s.Require().NoError(err)
}

func (s *ChangeApplierSuite) TestApplyRemote_UnknownChangeType_ReturnsError() {
	ctx := context.Background()
	rootName := domain.RootName("test")
	change := domain.RemoteChange{
		ChangeType: domain.SyncChangeType(999), // invalid
	}

	s.mockClientFactory.On("CurrentClient").Return(s.mockClient, nil)

	err := s.changeApplier.ApplyRemote(ctx, rootName, change)
	s.Require().ErrorIs(err, ErrUnknownChangeType)
}

func mustPath(t *testing.T, rawPath string) domain.Path {
	t.Helper()
	p, err := domain.NewPath(rawPath)
	require.NoError(t, err)
	return p
}

func mustScopedPath(t *testing.T, rootName domain.RootName, path domain.Path) domain.ScopedPath {
	t.Helper()
	sp, err := domain.NewScopedPath(rootName, path)
	require.NoError(t, err)
	return sp
}

// Mock implementations
type MockIFileManager struct {
	mock.Mock
}

func (m *MockIFileManager) RenameFile(ctx context.Context, oldScopedPath domain.ScopedPath, newScopedPath domain.ScopedPath) error {
	args := m.Called(ctx, oldScopedPath, newScopedPath)
	return args.Error(0)
}

func (m *MockIFileManager) DeleteFile(ctx context.Context, scopedPath domain.ScopedPath) error {
	args := m.Called(ctx, scopedPath)
	return args.Error(0)
}

func (m *MockIFileManager) PutFile(ctx context.Context, scopedPath domain.ScopedPath, content io.Reader) error {
	args := m.Called(ctx, scopedPath, content)
	return args.Error(0)
}

func (m *MockIFileManager) GetFile(ctx context.Context, scopedPath domain.ScopedPath) (io.ReadCloser, error) {
	args := m.Called(ctx, scopedPath)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

type MockIClientFactory struct {
	mock.Mock
}

func (m *MockIClientFactory) CurrentClient() (interfaces.IClient, error) {
	args := m.Called()
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

func (m *MockIClient) PutFile(ctx context.Context, content io.Reader, scopedPath domain.ScopedPath) error {
	args := m.Called(ctx, content, scopedPath)
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
