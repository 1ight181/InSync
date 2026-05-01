package filemanager

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	resource "insync/internal/infrastructure/filemanager/content"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"insync/internal/domain"
)

type FileManagerSuite struct {
	suite.Suite

	mockRootResolver   *MockIRootResolver
	mockHashManager    *MockIHashManager
	mockFileSystem     *MockIFileSystem
	mockPathTreeWriter *MockIPathTreeWriter
	fileManager        *FileManager
}

func (s *FileManagerSuite) SetupTest() {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	s.mockRootResolver = &MockIRootResolver{}
	s.mockHashManager = &MockIHashManager{}
	s.mockFileSystem = &MockIFileSystem{}
	s.mockPathTreeWriter = &MockIPathTreeWriter{}

	var err error
	s.fileManager, err = NewFileManager(FileManagerOptions{
		RootResolver:   s.mockRootResolver,
		HashManager:    s.mockHashManager,
		FileSystem:     s.mockFileSystem,
		PathTreeWriter: s.mockPathTreeWriter,
		Logger:         logger,
	})
	s.Require().NoError(err)
}

func (s *FileManagerSuite) TearDownTest() {
	s.mockRootResolver.AssertExpectations(s.T())
	s.mockHashManager.AssertExpectations(s.T())
	s.mockFileSystem.AssertExpectations(s.T())
	s.mockPathTreeWriter.AssertExpectations(s.T())
}

func TestFileManager(t *testing.T) {
	suite.Run(t, new(FileManagerSuite))
}

func (s *FileManagerSuite) TestNewFileManager_InvalidOpts_ReturnsError() {
	opts := FileManagerOptions{
		RootResolver:   s.mockRootResolver,
		HashManager:    s.mockHashManager,
		FileSystem:     s.mockFileSystem,
		PathTreeWriter: s.mockPathTreeWriter,
	}

	fm, err := NewFileManager(opts)
	s.ErrorIs(err, ErrInvalidOpts)
	s.Nil(fm)
}

func (s *FileManagerSuite) TestCreateDir_Success() {
	ctx := context.Background()
	rootName := domain.RootName("root")
	relativePath := mustPath(s.T(), "newdir")
	scopedPath := mustScopedPath(s.T(), rootName, relativePath)
	resolvedPath := mustPath(s.T(), filepath.Join(os.TempDir(), "newdir"))

	s.mockRootResolver.On("ResolveRoot", scopedPath).Return(resolvedPath, nil)
	s.mockFileSystem.On("Mkdir", resolvedPath, fs.FileMode(0755)).Return(nil)
	s.mockPathTreeWriter.On("AddPath", scopedPath).Return(nil)

	err := s.fileManager.CreateDir(ctx, scopedPath)
	s.Require().NoError(err)
}

func (s *FileManagerSuite) TestRenameFile_Success() {
	ctx := context.Background()
	rootName := domain.RootName("root")
	oldPath := mustPath(s.T(), "old.txt")
	newPath := mustPath(s.T(), "new.txt")
	oldScopedPath := mustScopedPath(s.T(), rootName, oldPath)
	newScopedPath := mustScopedPath(s.T(), rootName, newPath)
	oldResolved := mustPath(s.T(), filepath.Join(os.TempDir(), "old.txt"))
	newResolved := mustPath(s.T(), filepath.Join(os.TempDir(), "new.txt"))

	s.mockRootResolver.On("ResolveRoot", oldScopedPath).Return(oldResolved, nil)
	s.mockRootResolver.On("ResolveRoot", newScopedPath).Return(newResolved, nil)
	s.mockFileSystem.On("Rename", oldResolved, newResolved).Return(nil)
	s.mockPathTreeWriter.On("RemovePath", oldScopedPath).Return(nil)
	s.mockPathTreeWriter.On("AddPath", newScopedPath).Return(nil)
	s.mockHashManager.On("MarkDirty", mock.Anything, newScopedPath).Return(nil)

	err := s.fileManager.RenameFile(ctx, oldScopedPath, newScopedPath)
	s.Require().NoError(err)
}

func (s *FileManagerSuite) TestDeleteFile_Success() {
	ctx := context.Background()
	rootName := domain.RootName("root")
	scopedPath := mustScopedPath(s.T(), rootName, mustPath(s.T(), "file.txt"))
	resolved := mustPath(s.T(), filepath.Join(os.TempDir(), "file.txt"))

	s.mockRootResolver.On("ResolveRoot", scopedPath).Return(resolved, nil)
	s.mockFileSystem.On("Remove", resolved).Return(nil)
	s.mockPathTreeWriter.On("RemovePath", scopedPath).Return(nil)
	s.mockHashManager.On("MarkDirty", mock.Anything, scopedPath).Return(nil)

	err := s.fileManager.DeleteFile(ctx, scopedPath)
	s.Require().NoError(err)
}

func (s *FileManagerSuite) TestGetFile_Success() {
	ctx := context.Background()
	rootName := domain.RootName("root")
	scopedPath := mustScopedPath(s.T(), rootName, mustPath(s.T(), "file.txt"))
	resolved := mustPath(s.T(), filepath.Join(os.TempDir(), "file.txt"))
	testData := []byte("hello world")
	file := newMockFsFile(testData)

	s.mockRootResolver.On("ResolveRoot", scopedPath).Return(resolved, nil)
	s.mockFileSystem.On("Open", resolved).Return(file, nil)

	reader, err := s.fileManager.GetFile(ctx, scopedPath)
	s.Require().NoError(err)
	defer reader.Close()

	data, err := io.ReadAll(reader)
	s.Require().NoError(err)
	s.Equal(testData, data)
}

func (s *FileManagerSuite) TestPutFile_Success() {
	ctx := context.Background()
	rootName := domain.RootName("root")
	scopedPath := mustScopedPath(s.T(), rootName, mustPath(s.T(), "dir/file.txt"))
	resolved := mustPath(s.T(), filepath.Join(os.TempDir(), "dir", "file.txt"))
	testData := []byte("file content")

	s.mockRootResolver.On("ResolveRoot", scopedPath).Return(resolved, nil)
	s.mockFileSystem.On("AtomicWrite", resolved, mock.Anything).Return(nil)
	s.mockPathTreeWriter.On("AddPath", scopedPath).Return(nil)
	s.mockHashManager.On("MarkDirty", mock.Anything, scopedPath).Return(nil)

	err := s.fileManager.PutFile(ctx, scopedPath, bytes.NewReader(testData))
	s.Require().NoError(err)
}

func (s *FileManagerSuite) TestGetSnapshot_UsesResolveWithForceRecalc_WhenMetadataDiffers() {
	ctx := context.Background()
	rootName := domain.RootName("root")
	rootPath := mustPath(s.T(), os.TempDir())
	scopedPath := mustScopedPath(s.T(), rootName, mustPath(s.T(), "."))

	baseMetadata := domain.FileMetadata{
		ModifiedUnix: 1,
		SizeBytes:    0,
		IsDirectory:  true,
	}
	baseFileInfo, err := domain.NewFileInfo(baseMetadata, "base-hash")
	s.Require().NoError(err)
	baseEntry, err := domain.NewFileEntry(scopedPath.Path, 1, baseFileInfo)
	s.Require().NoError(err)
	snapshot := domain.NewSnapshot([]domain.FileEntry{baseEntry})
	baseSnapshot := domain.BaseSnapshot{
		Snapshot:  snapshot,
		IsInitial: false,
	}

	s.mockRootResolver.On("ResolveRoot", scopedPath).Return(rootPath, nil)
	s.mockFileSystem.On("ReadDir", rootPath).Return([]fs.DirEntry{}, nil)
	s.mockFileSystem.On("Stat", rootPath).Return(mockFileInfo{size: 0, modTime: time.Unix(2, 0), isDir: true}, nil)
	s.mockHashManager.On("ResolveWithForceRecalc", mock.Anything, mock.Anything, rootName).Return("recalc-hash", nil)

	snapshot, err = s.fileManager.GetSnapshot(ctx, rootName, &baseSnapshot)
	s.Require().NoError(err)
	s.Require().Len(snapshot.Files, 1)
	s.Equal("recalc-hash", snapshot.Files[0].FileInfo.Hash)
}

func (s *FileManagerSuite) TestGetSnapshot_UsesResolveHash_WhenMetadataMatches() {
	ctx := context.Background()
	rootName := domain.RootName("root")
	rootPath := mustPath(s.T(), os.TempDir())
	scopedPath := mustScopedPath(s.T(), rootName, mustPath(s.T(), "."))

	baseMetadata := domain.FileMetadata{
		ModifiedUnix: uint64(time.Now().Unix()),
		SizeBytes:    0,
		IsDirectory:  true,
	}
	baseFileInfo, err := domain.NewFileInfo(baseMetadata, "base-hash")
	s.Require().NoError(err)
	baseEntry, err := domain.NewFileEntry(scopedPath.Path, 1, baseFileInfo)
	s.Require().NoError(err)
	snapshot := domain.NewSnapshot([]domain.FileEntry{baseEntry})
	baseSnapshot := domain.BaseSnapshot{
		Snapshot:  snapshot,
		IsInitial: false,
	}

	s.mockRootResolver.On("ResolveRoot", scopedPath).Return(rootPath, nil)
	s.mockFileSystem.On("ReadDir", rootPath).Return([]fs.DirEntry{}, nil)
	s.mockFileSystem.On("Stat", rootPath).Return(mockFileInfo{size: 0, modTime: time.Unix(int64(baseMetadata.ModifiedUnix), 0), isDir: true}, nil)
	s.mockHashManager.On("ResolveHash", mock.Anything, mock.Anything, rootName).Return("cached-hash", nil)

	snapshot, err = s.fileManager.GetSnapshot(ctx, rootName, &baseSnapshot)
	s.Require().NoError(err)
	s.Require().Len(snapshot.Files, 1)
	s.Equal("cached-hash", snapshot.Files[0].FileInfo.Hash)
}

func mustPath(t *testing.T, raw string) domain.Path {
	t.Helper()
	p, err := domain.NewPath(raw)
	require.NoError(t, err)
	return p
}

func mustScopedPath(t *testing.T, rootName domain.RootName, path domain.Path) domain.ScopedPath {
	t.Helper()
	scopedPath, err := domain.NewScopedPath(rootName, path)
	require.NoError(t, err)
	return scopedPath
}

type MockIRootResolver struct {
	mock.Mock
}

func (m *MockIRootResolver) ResolveRoot(scopedPath domain.ScopedPath) (domain.Path, error) {
	args := m.Called(scopedPath)
	return args.Get(0).(domain.Path), args.Error(1)
}

type MockIHashManager struct {
	mock.Mock
}

func (m *MockIHashManager) ResolveHash(ctx context.Context, resourceContent resource.ResourceContent, rootName domain.RootName) (string, error) {
	args := m.Called(ctx, resourceContent, rootName)
	return args.String(0), args.Error(1)
}

func (m *MockIHashManager) ResolveWithForceRecalc(ctx context.Context, resourceContent resource.ResourceContent, rootName domain.RootName) (string, error) {
	args := m.Called(ctx, resourceContent, rootName)
	return args.String(0), args.Error(1)
}

func (m *MockIHashManager) MarkDirty(ctx context.Context, scopedPath domain.ScopedPath) error {
	args := m.Called(ctx, scopedPath)
	return args.Error(0)
}

type MockIFileSystem struct {
	mock.Mock
}

func (m *MockIFileSystem) ReadDir(fullPath domain.Path) ([]fs.DirEntry, error) {
	args := m.Called(fullPath)
	return args.Get(0).([]fs.DirEntry), args.Error(1)
}

func (m *MockIFileSystem) WalkDir(fullPath domain.Path, walkFn func(path string, d fs.DirEntry, err error) error) error {
	args := m.Called(fullPath, walkFn)
	return args.Error(0)
}

func (m *MockIFileSystem) Open(fullPath domain.Path) (fs.File, error) {
	args := m.Called(fullPath)
	return args.Get(0).(fs.File), args.Error(1)
}

func (m *MockIFileSystem) Remove(fullPath domain.Path) error {
	args := m.Called(fullPath)
	return args.Error(0)
}

func (m *MockIFileSystem) Rename(oldPath, newPath domain.Path) error {
	args := m.Called(oldPath, newPath)
	return args.Error(0)
}

func (m *MockIFileSystem) MkdirAll(path domain.Path, perm fs.FileMode) error {
	args := m.Called(path, perm)
	return args.Error(0)
}

func (m *MockIFileSystem) AtomicWrite(fullPath domain.Path, data io.Reader) error {
	args := m.Called(fullPath, data)
	return args.Error(0)
}

func (m *MockIFileSystem) Stat(fullPath domain.Path) (fs.FileInfo, error) {
	args := m.Called(fullPath)
	return args.Get(0).(fs.FileInfo), args.Error(1)
}

func (m *MockIFileSystem) Mkdir(fullPath domain.Path, perm fs.FileMode) error {
	args := m.Called(fullPath, perm)
	return args.Error(0)
}

type MockIPathTreeWriter struct {
	mock.Mock
}

func (m *MockIPathTreeWriter) AddPath(scopedPath domain.ScopedPath) error {
	args := m.Called(scopedPath)
	return args.Error(0)
}

func (m *MockIPathTreeWriter) RemovePath(scopedPath domain.ScopedPath) error {
	args := m.Called(scopedPath)
	return args.Error(0)
}

type mockFsFile struct {
	*bytes.Reader
	name string
	info fs.FileInfo
}

func newMockFsFile(data []byte) *mockFsFile {
	return &mockFsFile{
		Reader: bytes.NewReader(data),
		name:   "file.txt",
		info:   mockFileInfo{size: int64(len(data))},
	}
}

func (f *mockFsFile) Close() error {
	return nil
}

func (f *mockFsFile) Stat() (fs.FileInfo, error) {
	return f.info, nil
}

func (f *mockFsFile) Name() string {
	return f.name
}

type mockFileInfo struct {
	size    int64
	modTime time.Time
	isDir   bool
}

func (m mockFileInfo) Name() string       { return "file.txt" }
func (m mockFileInfo) Size() int64        { return m.size }
func (m mockFileInfo) Mode() fs.FileMode  { return 0 }
func (m mockFileInfo) ModTime() time.Time { return m.modTime }
func (m mockFileInfo) IsDir() bool        { return m.isDir }
func (m mockFileInfo) Sys() interface{}   { return nil }
