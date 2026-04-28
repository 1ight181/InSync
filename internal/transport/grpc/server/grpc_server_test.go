package server

import (
	"context"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
	"io"
	"log/slog"
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type MockFileUseCase struct {
	mock.Mock
}

func (m *MockFileUseCase) GetSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error) {
	args := m.Called(ctx, rootName)
	return args.Get(0).(domain.Snapshot), args.Error(1)
}

func (m *MockFileUseCase) DeleteFile(ctx context.Context, scopedPath domain.ScopedPath) error {
	args := m.Called(ctx, scopedPath)
	return args.Error(0)
}

func (m *MockFileUseCase) PutFile(ctx context.Context, scopedPath domain.ScopedPath, file io.Reader) error {
	args := m.Called(ctx, scopedPath, file)
	return args.Error(0)
}

func (m *MockFileUseCase) RenameFile(ctx context.Context, oldScopedPath domain.ScopedPath, newScopedPath domain.ScopedPath) error {
	args := m.Called(ctx, oldScopedPath, newScopedPath)
	return args.Error(0)
}

func (m *MockFileUseCase) GetFile(ctx context.Context, scopedPath domain.ScopedPath) (io.ReadCloser, error) {
	args := m.Called(ctx, scopedPath)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockFileUseCase) CreateDir(ctx context.Context, scopedPath domain.ScopedPath) error {
	args := m.Called(ctx, scopedPath)
	return args.Error(0)
}

type ServerSuite struct {
	suite.Suite
	mockFileUseCase *MockFileUseCase
	grpcServer      *GrpcServer
	server          *grpc.Server
	client          *grpc.ClientConn
	pbClient        insyncpb.FileSyncServiceClient
	lis             *bufconn.Listener
}

func (s *ServerSuite) SetupTest() {
	s.mockFileUseCase = &MockFileUseCase{}
	s.lis = bufconn.Listen(1024 * 1024)

	grpcServerOpts := GrpcServerOptions{
		FileUseCase:             s.mockFileUseCase,
		Creds:                   insecure.NewCredentials(),
		NetworkType:             "bufnet",
		Address:                 "bufnet",
		ServiceName:             "buf",
		ShouldStartHealthServer: false,
		ChunkSizeInBytes:        4096,
		Logger:                  slog.Default(),
		Listener:                s.lis,
	}

	grpcServer, err := NewGrpcServer(grpcServerOpts)
	s.Require().NoError(err)

	err = grpcServer.Start()
	s.Require().NoError(err)

	s.grpcServer = grpcServer

	withContextDialer := grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
		return s.lis.Dial()
	})

	wthTransportCredentials := grpc.WithTransportCredentials(insecure.NewCredentials())

	client, err := grpc.NewClient("passthrough:///bufnet", withContextDialer, wthTransportCredentials)
	s.Require().NoError(err)
	s.client = client
	s.pbClient = insyncpb.NewFileSyncServiceClient(s.client)
}

func (s *ServerSuite) TearDownTest() {
	if s.client != nil {
		_ = s.client.Close()
	}
	if s.grpcServer != nil {
		s.grpcServer.Stop(context.Background())
	}
	if s.lis != nil {
		s.lis.Close()
	}
}

func TestServerSuite(t *testing.T) {
	suite.Run(t, new(ServerSuite))
}

func (s *ServerSuite) TestServer_GetSnapshot_Success() {
	ctx := context.Background()
	rootName := domain.RootName("root")
	s.mockFileUseCase.On("GetSnapshot", mock.Anything, rootName).Return(domain.Snapshot{}, nil)
	responce, err := s.pbClient.GetSnapshot(ctx, &insyncpb.GetSnapshotRequest{
		RootName: rootName.String(),
	})
	s.Require().NoError(err)
	s.Require().NotNil(responce)

	s.mockFileUseCase.AssertExpectations(s.T())
}

func (s *ServerSuite) TestServer_DeleteFile_Success() {
	ctx := context.Background()
	rootName := domain.RootName("root")
	relativePath := domain.Path("file.txt")
	scopedPath := domain.ScopedPath{Root: rootName, Path: relativePath}
	s.mockFileUseCase.On("DeleteFile", mock.Anything, scopedPath).Return(nil)
	responce, err := s.pbClient.DeleteFile(ctx, &insyncpb.DeleteFileRequest{
		RootName:     rootName.String(),
		RelativePath: relativePath.String(),
	})
	s.Require().NoError(err)
	s.Require().NotNil(responce)

	s.mockFileUseCase.AssertExpectations(s.T())
}

func (s *ServerSuite) TestServer_RenameFile_Success() {
	ctx := context.Background()
	rootName := domain.RootName("root")
	oldRelativePath := domain.Path("old_file.txt")
	newRelativePath := domain.Path("new_file.txt")
	oldScopedPath := domain.ScopedPath{Root: rootName, Path: oldRelativePath}
	newScopedPath := domain.ScopedPath{Root: rootName, Path: newRelativePath}
	s.mockFileUseCase.On("RenameFile", mock.Anything, oldScopedPath, newScopedPath).Return(nil)
	responce, err := s.pbClient.RenameFile(ctx, &insyncpb.RenameFileRequest{
		RootName:        rootName.String(),
		OldRelativePath: oldRelativePath.String(),
		NewRelativePath: newRelativePath.String(),
	})
	s.Require().NoError(err)
	s.Require().NotNil(responce)

	s.mockFileUseCase.AssertExpectations(s.T())
}

func (s *ServerSuite) TestServer_GetFile_Success() {
	ctx := context.Background()
	rootName := domain.RootName("root")
	relativePath := domain.Path("file.txt")
	scopedPath := domain.ScopedPath{Root: rootName, Path: relativePath}
	fileContent := "file content"
	s.mockFileUseCase.On("GetFile", mock.Anything, scopedPath).Return(io.NopCloser(strings.NewReader(fileContent)), nil)
	stream, err := s.pbClient.GetFile(ctx, &insyncpb.GetFileRequest{
		RootName:     rootName.String(),
		RelativePath: relativePath.String(),
	})
	s.Require().NoError(err)
	s.Require().NotNil(stream)

	var received []byte
	expected := []byte(fileContent)
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(s.T(), err)

		received = append(received, msg.GetChunk().GetData()...)
	}

	require.Equal(s.T(), expected, received)
	s.mockFileUseCase.AssertExpectations(s.T())
}

func (s *ServerSuite) TestServer_PutFile_Success() {
	ctx := context.Background()
	rootName := domain.RootName("root")
	relativePath := domain.Path("file.txt")
	scopedPath := domain.ScopedPath{Root: rootName, Path: relativePath}
	fileContent := "file content"

	done := make(chan struct{})

	s.mockFileUseCase.On("PutFile", mock.Anything, scopedPath, mock.Anything).
		Run(func(args mock.Arguments) {
			reader := args.Get(2).(io.Reader)

			go func() {
				defer close(done)

				data, err := io.ReadAll(reader)
				if err != nil {
					s.T().Logf("Warning: failed to read from pipe: %v", err)
					return
				}

				s.Require().Equal(fileContent, string(data))
			}()
		}).
		Return(nil).
		Once()

	stream, err := s.pbClient.PutFile(ctx)
	s.Require().NoError(err)

	err = stream.Send(&insyncpb.PutFileRequest{
		Payload: &insyncpb.PutFileRequest_Init{
			Init: &insyncpb.PutFileInit{
				RootName:     rootName.String(),
				RelativePath: relativePath.String(),
			},
		},
	})
	s.Require().NoError(err)

	err = stream.Send(&insyncpb.PutFileRequest{
		Payload: &insyncpb.PutFileRequest_Chunk{
			Chunk: &insyncpb.FileChunk{
				Data: []byte(fileContent),
			},
		},
	})
	s.Require().NoError(err)

	err = stream.CloseSend()
	s.Require().NoError(err)

	_, err = stream.CloseAndRecv()
	s.Require().NoError(err)

	<-done

	s.mockFileUseCase.AssertExpectations(s.T())
}
func Test_domainFileEntryToPb(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		domainEntry domain.FileEntry
		expectedPb  *insyncpb.FileEntry
	}{
		{
			name: "full mapping",
			domainEntry: domain.FileEntry{
				RelativePath: "dir/file.txt",
				SubtreeSize:  42,
				FileInfo: domain.FileInfo{
					Hash: "abc123",
					Metadata: domain.FileMetadata{
						SizeBytes:    100,
						ModifiedUnix: 123456789,
						IsDirectory:  false,
					},
				},
			},
			expectedPb: &insyncpb.FileEntry{
				RelativePath: "dir/file.txt",
				SizeBytes:    100,
				ModifiedUnix: 123456789,
				Hash:         "abc123",
				IsDirectory:  false,
				SubtreeSize:  42,
			},
		},
		{
			name: "directory entry",
			domainEntry: domain.FileEntry{
				RelativePath: "dir",
				SubtreeSize:  10,
				FileInfo: domain.FileInfo{
					Hash: "",
					Metadata: domain.FileMetadata{
						SizeBytes:    0,
						ModifiedUnix: 111,
						IsDirectory:  true,
					},
				},
			},
			expectedPb: &insyncpb.FileEntry{
				RelativePath: "dir",
				SizeBytes:    0,
				ModifiedUnix: 111,
				Hash:         "",
				IsDirectory:  true,
				SubtreeSize:  10,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := domainFileEntryToPb(tc.domainEntry)

			require.Equal(t, tc.expectedPb, result)
		})
	}
}

func Test_domainFileEntriesToPb(t *testing.T) {
	t.Parallel()

	input := []domain.FileEntry{
		{
			RelativePath: "b.txt",
			SubtreeSize:  2,
			FileInfo: domain.FileInfo{
				Hash: "hash2",
				Metadata: domain.FileMetadata{
					SizeBytes:    20,
					ModifiedUnix: 200,
					IsDirectory:  false,
				},
			},
		},
		{
			RelativePath: "a.txt",
			SubtreeSize:  1,
			FileInfo: domain.FileInfo{
				Hash: "hash1",
				Metadata: domain.FileMetadata{
					SizeBytes:    10,
					ModifiedUnix: 100,
					IsDirectory:  false,
				},
			},
		},
	}

	expected := []*insyncpb.FileEntry{
		{
			RelativePath: "b.txt",
			SizeBytes:    20,
			ModifiedUnix: 200,
			Hash:         "hash2",
			IsDirectory:  false,
			SubtreeSize:  2,
		},
		{
			RelativePath: "a.txt",
			SizeBytes:    10,
			ModifiedUnix: 100,
			Hash:         "hash1",
			IsDirectory:  false,
			SubtreeSize:  1,
		},
	}

	result := domainFileEntriesToPb(input)

	require.Equal(t, expected, result)
}

func Test_domainSnapshotToPb(t *testing.T) {
	t.Parallel()

	snapshot := domain.Snapshot{
		Files: []domain.FileEntry{
			{
				RelativePath: "file.txt",
				SubtreeSize:  1,
				FileInfo: domain.FileInfo{
					Hash: "h1",
					Metadata: domain.FileMetadata{
						SizeBytes:    10,
						ModifiedUnix: 111,
						IsDirectory:  false,
					},
				},
			},
		},
	}

	result := domainSnapshotToPb(snapshot)

	require.Len(t, result.Files, 1)
	require.Equal(t, "file.txt", result.Files[0].RelativePath)
	require.Equal(t, "h1", result.Files[0].Hash)
}

func (s *ServerSuite) TestServer_CreateDir_Success() {
	ctx := context.Background()
	rootName := domain.RootName("root")
	relativePath := domain.Path("new_dir")
	s.mockFileUseCase.On("CreateDir", mock.Anything, domain.ScopedPath{Root: rootName, Path: relativePath}).Return(nil).Once()
	responce, err := s.pbClient.CreateDir(ctx, &insyncpb.CreateDirRequest{
		RootName:     rootName.String(),
		RelativePath: relativePath.String(),
	})
	s.Require().NoError(err)
	s.Require().NotNil(responce)

	s.mockFileUseCase.AssertExpectations(s.T())
}
