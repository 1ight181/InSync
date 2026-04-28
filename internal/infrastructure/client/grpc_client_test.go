package client

import (
	"bytes"
	"context"
	"fmt"
	"insync/internal/transport/grpc/insyncpb"
	"io"
	"log/slog"
	"net"
	"testing"

	domain "insync/internal/domain"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ClientSuite struct {
	suite.Suite

	client *GrpcClient
	lis    *bufconn.Listener
	server *grpc.Server
}

func (s *ClientSuite) SetupTest() {
	s.lis = bufconn.Listen(1024)
	s.server = grpc.NewServer()
	insyncpb.RegisterFileSyncServiceServer(s.server, &testClient_FileSyncServer{})

	go s.server.Serve(s.lis)

	conf := GrpcConf{
		DefaultServiceConf: `{}`,
		ResolverScheme:     "passthrough",
		Creds:              insecure.NewCredentials(),
		Dialer: func(ctx context.Context, _ string) (net.Conn, error) {
			return s.lis.Dial()
		},
		ServerAddress:    "/bufnet",
		ChunkSizeInBytes: 4096,
	}

	clientOpts := GrpcClientOptions{
		Conf:   &conf,
		Logger: slog.Default(),
	}

	client, err := NewGrpcClient(clientOpts)
	s.client = client
	s.Require().NoError(err)
}

func (s *ClientSuite) TearDownTest() {
	if s.client != nil && s.client.isStarted.Load() {
		_ = s.client.Close()
	}
	if s.server != nil {
		s.server.Stop()
	}
	if s.lis != nil {
		s.lis.Close()
	}
}

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}

func (s *ClientSuite) TestClient_Connect_Success() {
	err := s.client.Connect()
	s.Require().NoError(err)
	s.True(s.client.isStarted.Load())
}

func (s *ClientSuite) TestClient_GetSnapshot_Success() {
	s.Require().NoError(s.client.Connect())

	snapshot, err := s.client.GetSnapshot(context.Background(), domain.RootName("photos"))
	s.Require().NoError(err)
	s.NotEmpty(snapshot.Files)
}

func (s *ClientSuite) TestClient_GetFile_Success() {
	s.Require().NoError(s.client.Connect())

	scopedPath, err := domain.NewScopedPath(domain.RootName("photos"), domain.Path("chunk0chunk1chunk2"))
	s.Require().NoError(err)

	reader, err := s.client.GetFile(context.Background(), scopedPath)
	s.Require().NoError(err)
	defer reader.Close()

	data, err := io.ReadAll(reader)
	s.Require().NoError(err)
	s.Equal("chunk0chunk1chunk2", string(data))
}

func (s *ClientSuite) TestClient_PutFile_Success() {
	s.Require().NoError(s.client.Connect())

	data := []byte("testClient_ file content from client")
	scopedPath, err := domain.NewScopedPath(domain.RootName("photos"), domain.Path("new.txt"))
	s.Require().NoError(err)
	err = s.client.PutFile(context.Background(), bytes.NewReader(data), scopedPath)
	s.Require().NoError(err)
}

func (s *ClientSuite) TestClient_Delete_Success() {
	s.Require().NoError(s.client.Connect())

	scopedPath, err := domain.NewScopedPath(domain.RootName("photos"), domain.Path("old.txt"))
	s.Require().NoError(err)
	err = s.client.DeleteFile(context.Background(), scopedPath)
	s.Require().NoError(err)
}

func (s *ClientSuite) TestClient_Rename_Success() {
	s.Require().NoError(s.client.Connect())

	oldScopedPath, err := domain.NewScopedPath(domain.RootName("photos"), domain.Path("old.txt"))
	s.Require().NoError(err)
	newScopedPath, err := domain.NewScopedPath(domain.RootName("photos"), domain.Path("new.txt"))
	s.Require().NoError(err)
	err = s.client.RenameFile(context.Background(), oldScopedPath, newScopedPath)
	s.Require().NoError(err)
}

func (s *ClientSuite) TestClient_CreateDir_Success() {
	s.Require().NoError(s.client.Connect())

	scopedPath, err := domain.NewScopedPath(domain.RootName("photos"), domain.Path("newdir"))
	s.Require().NoError(err)
	err = s.client.CreateDir(context.Background(), scopedPath)
	s.Require().NoError(err)
}

func (s *ClientSuite) TestClient_Methods_WhenNotStarted_ReturnErrClientNotStarted() {
	s.client.isStarted.Store(false)
	_, err := s.client.GetSnapshot(context.Background(), domain.RootName("root"))
	s.Require().ErrorIs(err, ErrClientNotStarted)

	scopedPath, err := domain.NewScopedPath(domain.RootName("root"), s.mustNewPath("f"))
	s.Require().NoError(err)
	_, err = s.client.GetFile(context.Background(), scopedPath)
	s.Require().ErrorIs(err, ErrClientNotStarted)

	err = s.client.PutFile(context.Background(), bytes.NewReader(nil), scopedPath)
	s.Require().ErrorIs(err, ErrClientNotStarted)

	err = s.client.DeleteFile(context.Background(), scopedPath)
	s.Require().ErrorIs(err, ErrClientNotStarted)

	err = s.client.RenameFile(context.Background(), scopedPath, scopedPath)
	s.Require().ErrorIs(err, ErrClientNotStarted)

	err = s.client.CreateDir(context.Background(), scopedPath)
	s.Require().ErrorIs(err, ErrClientNotStarted)
}

func (s *ClientSuite) mustNewPath(path string) domain.Path {
	s.T().Helper()
	p, err := domain.NewPath(path)
	s.Require().NoError(err)
	return p
}

type testClient_FileSyncServer struct {
	insyncpb.UnimplementedFileSyncServiceServer
}

func (s *testClient_FileSyncServer) PutFile(stream grpc.ClientStreamingServer[insyncpb.PutFileRequest, emptypb.Empty]) error {
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&emptypb.Empty{})
		}
		if err != nil {
			return err
		}
	}
}

func (s *testClient_FileSyncServer) GetSnapshot(_ context.Context, _ *insyncpb.GetSnapshotRequest) (*insyncpb.GetSnapshotResponse, error) {
	return &insyncpb.GetSnapshotResponse{
		Snapshot: &insyncpb.Snapshot{
			Files: []*insyncpb.FileEntry{
				{RelativePath: "testClient_/file.txt", ModifiedUnix: 1234567890, SizeBytes: 42, Hash: "abc123"},
			},
		},
	}, nil
}

func (s *testClient_FileSyncServer) GetFile(_ *insyncpb.GetFileRequest, stream insyncpb.FileSyncService_GetFileServer) error {
	for i := 0; i < 3; i++ {
		_ = stream.Send(&insyncpb.GetFileResponse{
			Chunk: &insyncpb.FileChunk{Index: uint64(i), Data: []byte(fmt.Sprintf("chunk%d", i))},
		})
	}
	return nil
}

func (s *testClient_FileSyncServer) DeleteFile(context.Context, *insyncpb.DeleteFileRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *testClient_FileSyncServer) RenameFile(context.Context, *insyncpb.RenameFileRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *testClient_FileSyncServer) CreateDir(context.Context, *insyncpb.CreateDirRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
