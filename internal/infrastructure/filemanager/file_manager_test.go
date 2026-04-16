package filemanager

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"insync/internal/domain"
	cont "insync/internal/infrastructure/filemanager/content"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type tempReadWriteCloser struct {
	bytes.Buffer
	closed bool
}

func (t *tempReadWriteCloser) Close() error {
	t.closed = true
	return nil
}

type readOnlyFile struct {
	io.Reader
	info   fs.FileInfo
	closed bool
}

func (r *readOnlyFile) Read(p []byte) (int, error) { return r.Reader.Read(p) }
func (r *readOnlyFile) Close() error               { r.closed = true; return nil }
func (r *readOnlyFile) Stat() (fs.FileInfo, error) { return r.info, nil }

type failingReadCloser struct{ err error }

func (f failingReadCloser) Read([]byte) (int, error) { return 0, f.err }
func (f failingReadCloser) Close() error             { return nil }

type stubCloser struct {
	closeErr error
	closed   bool
}

func (s *stubCloser) Close() error {
	s.closed = true
	return s.closeErr
}

type stubFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

func (s stubFileInfo) Name() string       { return s.name }
func (s stubFileInfo) Size() int64        { return s.size }
func (s stubFileInfo) Mode() fs.FileMode  { return s.mode }
func (s stubFileInfo) ModTime() time.Time { return s.modTime }
func (s stubFileInfo) IsDir() bool        { return s.isDir }
func (s stubFileInfo) Sys() any           { return nil }

type stubDirEntry struct {
	name     string
	isDir    bool
	symlink  bool
	contents string
	infoErr  error
}

func (s stubDirEntry) Name() string { return s.name }

func (s stubDirEntry) IsDir() bool { return s.isDir }

func (s stubDirEntry) Type() fs.FileMode {
	if s.symlink {
		return fs.ModeSymlink
	}
	if s.isDir {
		return fs.ModeDir
	}
	return 0
}

func (s stubDirEntry) Info() (fs.FileInfo, error) {
	if s.infoErr != nil {
		return nil, s.infoErr
	}
	return stubFileInfo{
		name:    s.name,
		size:    int64(len(s.contents)),
		modTime: time.Unix(100, 0),
		isDir:   s.isDir,
	}, nil
}

type fakeFileSystem struct {
	readDirFunc    func(fullPath string) ([]fs.DirEntry, error)
	walkDirFunc    func(fullPath string, walkFn func(path string, d fs.DirEntry, err error) error) error
	openFunc       func(fullPath string) (fs.File, error)
	removeFunc     func(fullPath string) error
	createTempFunc func(dir, pattern string) (io.ReadWriteCloser, string, error)
	renameFunc     func(oldPath, newPath string) error
	mkdirAllFunc   func(path string, perm fs.FileMode) error
	createFunc     func(fullPath string) (io.ReadWriteCloser, error)
	statFunc       func(fullPath string) (fs.FileInfo, error)

	mu              sync.Mutex
	removedPaths    []string
	renamedPairs    [][2]string
	mkdirAllPaths   []string
	createdPaths    []string
	openedPaths     []string
	readDirPaths    []string
	statPaths       []string
	createTempPaths []string
}

func (f *fakeFileSystem) ReadDir(fullPath string) ([]fs.DirEntry, error) {
	f.mu.Lock()
	f.readDirPaths = append(f.readDirPaths, fullPath)
	f.mu.Unlock()
	if f.readDirFunc != nil {
		return f.readDirFunc(fullPath)
	}
	return nil, nil
}

func (f *fakeFileSystem) WalkDir(fullPath string, walkFn func(path string, d fs.DirEntry, err error) error) error {
	if f.walkDirFunc != nil {
		return f.walkDirFunc(fullPath, walkFn)
	}
	return nil
}

func (f *fakeFileSystem) Open(fullPath string) (fs.File, error) {
	f.mu.Lock()
	f.openedPaths = append(f.openedPaths, fullPath)
	f.mu.Unlock()
	if f.openFunc != nil {
		return f.openFunc(fullPath)
	}
	return nil, fmt.Errorf("unexpected open: %s", fullPath)
}

func (f *fakeFileSystem) Remove(fullPath string) error {
	f.mu.Lock()
	f.removedPaths = append(f.removedPaths, fullPath)
	f.mu.Unlock()
	if f.removeFunc != nil {
		return f.removeFunc(fullPath)
	}
	return nil
}

func (f *fakeFileSystem) CreateTempFile(dir, pattern string) (io.ReadWriteCloser, string, error) {
	f.mu.Lock()
	f.createTempPaths = append(f.createTempPaths, filepath.Join(dir, pattern))
	f.mu.Unlock()
	if f.createTempFunc != nil {
		return f.createTempFunc(dir, pattern)
	}
	return &tempReadWriteCloser{}, filepath.Join(dir, "temp-file"), nil
}

func (f *fakeFileSystem) Rename(oldPath, newPath string) error {
	f.mu.Lock()
	f.renamedPairs = append(f.renamedPairs, [2]string{oldPath, newPath})
	f.mu.Unlock()
	if f.renameFunc != nil {
		return f.renameFunc(oldPath, newPath)
	}
	return nil
}

func (f *fakeFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	f.mu.Lock()
	f.mkdirAllPaths = append(f.mkdirAllPaths, path)
	f.mu.Unlock()
	if f.mkdirAllFunc != nil {
		return f.mkdirAllFunc(path, perm)
	}
	return nil
}

func (f *fakeFileSystem) Create(fullPath string) (io.ReadWriteCloser, error) {
	f.mu.Lock()
	f.createdPaths = append(f.createdPaths, fullPath)
	f.mu.Unlock()
	if f.createFunc != nil {
		return f.createFunc(fullPath)
	}
	return &tempReadWriteCloser{}, nil
}

func (f *fakeFileSystem) Stat(fullPath string) (fs.FileInfo, error) {
	f.mu.Lock()
	f.statPaths = append(f.statPaths, fullPath)
	f.mu.Unlock()
	if f.statFunc != nil {
		return f.statFunc(fullPath)
	}
	return stubFileInfo{name: filepath.Base(fullPath), modTime: time.Unix(100, 0)}, nil
}

type stubRootResolver struct {
	resolvedPaths []string
	calls         int
	mu            sync.Mutex
}

func (s *stubRootResolver) ResolveRoot(_ domain.RootName, _ domain.Path) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.calls >= len(s.resolvedPaths) {
		return "", fmt.Errorf("unexpected ResolveRoot call %d", s.calls)
	}
	resolvedPath := s.resolvedPaths[s.calls]
	s.calls++
	return resolvedPath, nil
}

type stubHashManager struct {
	resolveHashFunc func(resourceContent cont.ResourceContent, fullPath string) (string, error)
	markDirtyFunc   func(fullPath string) error
	resolveCalls    []string
	markDirtyCalls  []string
}

func (s *stubHashManager) ResolveHash(resourceContent cont.ResourceContent, fullPath string) (string, error) {
	s.resolveCalls = append(s.resolveCalls, fullPath)
	if s.resolveHashFunc != nil {
		return s.resolveHashFunc(resourceContent, fullPath)
	}
	return "hash-value", nil
}

func (s *stubHashManager) MarkDirty(fullPath string) error {
	s.markDirtyCalls = append(s.markDirtyCalls, fullPath)
	if s.markDirtyFunc != nil {
		return s.markDirtyFunc(fullPath)
	}
	return nil
}

type stubPathTreeWriter struct {
	added   []string
	removed []string
	aErr    error
	rErr    error
}

func (s *stubPathTreeWriter) AddPath(fullPath string) error {
	s.added = append(s.added, fullPath)
	return s.aErr
}

func (s *stubPathTreeWriter) RemovePath(fullPath string) error {
	s.removed = append(s.removed, fullPath)
	return s.rErr
}

func newTestManager(t *testing.T, fsImpl *fakeFileSystem, resolver *stubRootResolver) (*FileManager, *stubHashManager, *stubPathTreeWriter) {
	t.Helper()
	if fsImpl == nil {
		fsImpl = &fakeFileSystem{}
	}
	if resolver == nil {
		resolver = &stubRootResolver{}
	}

	hashManager := &stubHashManager{}
	pathTreeWriter := &stubPathTreeWriter{}

	manager := NewFileManager(FileManagerOptions{
		RootResolver:   resolver,
		HashManager:    hashManager,
		FileSystem:     fsImpl,
		PathTreeWriter: pathTreeWriter,
		TempDir:        t.TempDir(),
		Logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
		LoggerCtx:      context.Background(),
	})

	return manager, hashManager, pathTreeWriter
}

func TestNewMultiCloser_CollectsErrorsFromAllClosers(t *testing.T) {
	firstCloser := &stubCloser{closeErr: errors.New("first")}
	secondCloser := &stubCloser{closeErr: errors.New("second")}

	err := NewMultiCloser(firstCloser, secondCloser).Close()

	require.Error(t, err)
	assert.True(t, firstCloser.closed)
	assert.True(t, secondCloser.closed)
	assert.Contains(t, err.Error(), "first")
	assert.Contains(t, err.Error(), "second")
}

func TestNewFileManager_PanicsWhenOptionsMissing(t *testing.T) {
	require.Panics(t, func() {
		NewFileManager(FileManagerOptions{})
	})
}

func TestWriteContentToTemp_RemovesTempFileOnCopyError(t *testing.T) {
	filesystem := &fakeFileSystem{}
	manager, _, _ := newTestManager(t, filesystem, nil)

	filesystem.createTempFunc = func(dir, pattern string) (io.ReadWriteCloser, string, error) {
		return &tempReadWriteCloser{}, filepath.Join(dir, "temp-file.txt"), nil
	}

	tempPath, err := manager.writeContentToTemp(failingReadCloser{err: errors.New("copy failed")})

	require.Error(t, err)
	assert.Empty(t, tempPath)
	assert.Equal(t,
		[]string{filepath.Join(manager.tempDir, "temp-file.txt")},
		filesystem.removedPaths,
	)
}

func TestMoveTempToDestination_RenamesAndRemovesTempFile(t *testing.T) {
	filesystem := &fakeFileSystem{}
	manager, _, _ := newTestManager(t, filesystem, nil)

	tempFilePath := filepath.Join(manager.tempDir, "temp-file.txt")
	destinationPath := filepath.Join(manager.tempDir, "nested", "file.txt")

	err := manager.moveTempToDestination(context.Background(), tempFilePath, destinationPath)

	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Dir(destinationPath)}, filesystem.mkdirAllPaths)
	assert.Equal(t, [][2]string{{tempFilePath, destinationPath}}, filesystem.renamedPairs)
	assert.Equal(t, []string{tempFilePath}, filesystem.removedPaths)
}

func TestMoveTempToDestination_FallsBackOnEXDEV(t *testing.T) {
	filesystem := &fakeFileSystem{}
	manager, _, _ := newTestManager(t, filesystem, nil)

	tempFilePath := filepath.Join(manager.tempDir, "temp-file.txt")
	destinationPath := filepath.Join(manager.tempDir, "nested", "file.txt")
	sourceContent := []byte("payload")
	destinationBuffer := &tempReadWriteCloser{}

	filesystem.renameFunc = func(oldPath, newPath string) error { return syscall.EXDEV }
	filesystem.openFunc = func(fullPath string) (fs.File, error) {
		require.Equal(t, tempFilePath, fullPath)
		return &readOnlyFile{
			Reader: bytes.NewReader(sourceContent),
			info:   stubFileInfo{name: filepath.Base(fullPath)},
		}, nil
	}
	filesystem.createFunc = func(fullPath string) (io.ReadWriteCloser, error) {
		require.Equal(t, destinationPath, fullPath)
		return destinationBuffer, nil
	}

	err := manager.moveTempToDestination(context.Background(), tempFilePath, destinationPath)

	require.NoError(t, err)
	assert.Equal(t, string(sourceContent), destinationBuffer.String())
	assert.Equal(t, []string{tempFilePath}, filesystem.openedPaths)
	assert.Equal(t, []string{destinationPath}, filesystem.createdPaths)
	assert.Equal(t, []string{tempFilePath}, filesystem.removedPaths)
}

func TestMoveTempToDestination_ReturnsContextErrorDuringFallback(t *testing.T) {
	filesystem := &fakeFileSystem{}
	manager, _, _ := newTestManager(t, filesystem, nil)

	tempFilePath := filepath.Join(manager.tempDir, "temp-file.txt")
	destinationPath := filepath.Join(manager.tempDir, "nested", "file.txt")

	filesystem.renameFunc = func(oldPath, newPath string) error { return syscall.EXDEV }

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := manager.moveTempToDestination(ctx, tempFilePath, destinationPath)

	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, []string{tempFilePath}, filesystem.removedPaths)
}

func TestOpenFileContentWithHeader_PrefixesRelativePath(t *testing.T) {
	filesystem := &fakeFileSystem{}
	manager, _, _ := newTestManager(t, filesystem, nil)

	fullPath := filepath.Join(manager.tempDir, "file.txt")
	relativePath := "dir/file.txt"

	filesystem.openFunc = func(path string) (fs.File, error) {
		require.Equal(t, fullPath, path)
		return &readOnlyFile{
			Reader: bytes.NewReader([]byte("body")),
			info:   stubFileInfo{name: "file.txt"},
		}, nil
	}

	content, err := manager.openFileContentWithHeader(fullPath, relativePath)
	require.NoError(t, err)
	defer content.Close()

	buffer, err := io.ReadAll(content)
	require.NoError(t, err)

	assert.Equal(t, relativePath+"body", string(buffer))
}

func TestOpenDirContent_SortsEntriesAndSkipsSymlinks(t *testing.T) {
	filesystem := &fakeFileSystem{}
	manager, _, _ := newTestManager(t, filesystem, nil)

	fullPath := filepath.Join(manager.tempDir, "root")

	filesystem.readDirFunc = func(path string) ([]fs.DirEntry, error) {
		require.Equal(t, fullPath, path)
		return []fs.DirEntry{
			stubDirEntry{name: "b.txt", contents: "B"},
			stubDirEntry{name: "link", symlink: true},
			stubDirEntry{name: "a.txt", contents: "A"},
		}, nil
	}

	filesystem.openFunc = func(path string) (fs.File, error) {
		switch filepath.Base(path) {
		case "a.txt":
			return &readOnlyFile{Reader: bytes.NewReader([]byte("A")), info: stubFileInfo{name: "a.txt"}}, nil
		case "b.txt":
			return &readOnlyFile{Reader: bytes.NewReader([]byte("B")), info: stubFileInfo{name: "b.txt"}}, nil
		default:
			return nil, fmt.Errorf("unexpected open path: %s", path)
		}
	}

	content, err := manager.openDirContent(fullPath, "root")
	require.NoError(t, err)
	defer content.Close()

	buffer, err := io.ReadAll(content)
	require.NoError(t, err)

	expected := "root" +
		filepath.Join("root", "a.txt") + "A" +
		filepath.Join("root", "b.txt") + "B"

	assert.Equal(t, expected, string(buffer))
	assert.Equal(t,
		[]string{
			filepath.Join(fullPath, "a.txt"),
			filepath.Join(fullPath, "b.txt"),
		},
		filesystem.openedPaths,
	)
}
