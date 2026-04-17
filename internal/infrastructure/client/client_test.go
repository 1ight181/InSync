package client

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
	"io"
	"log/slog"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func mustRootName(t *testing.T, value string) domain.RootName {
	t.Helper()

	rootName, err := domain.NewRootName(value)
	if err != nil {
		t.Fatalf("NewRootName(%q) returned error: %v", value, err)
	}

	return rootName
}

func mustPath(t *testing.T, value string) domain.Path {
	t.Helper()

	pathValue, err := domain.NewPath(value)
	if err != nil {
		t.Fatalf("NewPath(%q) returned error: %v", value, err)
	}

	return pathValue
}

func baseGrpcConf() *GrpcConf {
	return &GrpcConf{
		CertPath:             "/tmp/client.crt",
		KeyPath:              "/tmp/client.key",
		CaCertPath:           "/tmp/ca.crt",
		ServerNetworkType:    "tcp",
		ServerAddress:        "127.0.0.1:12345",
		ServerServiceName:    "filesync",
		ServerName:           "localhost",
		ResolverScheme:       "dns",
		LoadBalancingPolicy:  "round_robin",
		ShouldUseHealthCheck: true,
		RpcTimeout:           5 * time.Second,
		RetryPolicy: &RpcRetryPolicy{
			MaxAttempts:          3,
			InitialBackoff:       100 * time.Millisecond,
			MaxBackoff:           2 * time.Second,
			BackoffMultiplier:    1.5,
			RetryableStatusCodes: []string{"UNAVAILABLE", "DEADLINE_EXCEEDED"},
		},
		ConnectionConfig: &ConnectionConfig{
			BaseDelay:         100 * time.Millisecond,
			Multiplier:        1.5,
			MaxDelay:          2 * time.Second,
			Jitter:            0.2,
			MinConnectTimeout: 3 * time.Second,
		},
		ChunkSizeInBytes: 4,
	}
}

func startedClientWithFakeService(t *testing.T, fakeClient insyncpb.FileSyncServiceClient) *GrpcClient {
	t.Helper()

	grpcClient := &GrpcClient{
		conf:      baseGrpcConf(),
		ctx:       context.Background(),
		logger:    testLogger(),
		loggerCtx: context.Background(),
		client:    fakeClient,
	}
	grpcClient.isStarted.Store(true)

	return grpcClient
}

func TestNewGrpcClient_PanicsWhenConfigIsIncomplete(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name   string
		mutate func(conf *GrpcConf)
	}

	cases := []testCase{
		{
			name: "nil config",
			mutate: func(conf *GrpcConf) {
				// handled separately
			},
		},
		{
			name:   "empty cert path",
			mutate: func(conf *GrpcConf) { conf.CertPath = "" },
		},
		{
			name:   "empty key path",
			mutate: func(conf *GrpcConf) { conf.KeyPath = "" },
		},
		{
			name:   "empty ca path",
			mutate: func(conf *GrpcConf) { conf.CaCertPath = "" },
		},
		{
			name:   "empty network type",
			mutate: func(conf *GrpcConf) { conf.ServerNetworkType = "" },
		},
		{
			name:   "empty address",
			mutate: func(conf *GrpcConf) { conf.ServerAddress = "" },
		},
		{
			name:   "empty service name",
			mutate: func(conf *GrpcConf) { conf.ServerServiceName = "" },
		},
		{
			name:   "empty server name",
			mutate: func(conf *GrpcConf) { conf.ServerName = "" },
		},
		{
			name:   "empty resolver scheme",
			mutate: func(conf *GrpcConf) { conf.ResolverScheme = "" },
		},
		{
			name:   "empty load balancing policy",
			mutate: func(conf *GrpcConf) { conf.LoadBalancingPolicy = "" },
		},
		{
			name:   "nil retry policy",
			mutate: func(conf *GrpcConf) { conf.RetryPolicy = nil },
		},
		{
			name:   "nil connection config",
			mutate: func(conf *GrpcConf) { conf.ConnectionConfig = nil },
		},
		{
			name:   "zero chunk size",
			mutate: func(conf *GrpcConf) { conf.ChunkSizeInBytes = 0 },
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var conf *GrpcConf
			if tc.name != "nil config" {
				conf = baseGrpcConf()
				tc.mutate(conf)
			}

			defer func() {
				recovered := recover()
				if recovered == nil {
					t.Fatalf("expected panic")
				}
			}()

			_ = NewGrpcClient(GrpcClientOptions{
				Conf:   conf,
				Ctx:    context.Background(),
				Logger: testLogger(),
			})
		})
	}
}

func TestNewGrpcClient_ReturnsClient(t *testing.T) {
	t.Parallel()

	grpcClient := NewGrpcClient(GrpcClientOptions{
		Conf:   baseGrpcConf(),
		Ctx:    context.Background(),
		Logger: testLogger(),
	})

	if grpcClient == nil {
		t.Fatalf("expected client, got nil")
	}

	if grpcClient.conf == nil {
		t.Fatalf("expected config to be stored")
	}
}

func TestCreateAddress(t *testing.T) {
	t.Parallel()

	grpcClient := &GrpcClient{conf: baseGrpcConf()}

	gotAddress := grpcClient.createAddress()
	wantAddress := "dns://127.0.0.1:12345"

	if gotAddress != wantAddress {
		t.Fatalf("unexpected address: got %q, want %q", gotAddress, wantAddress)
	}
}

func TestCreateServiceConfig_WithHealthCheck(t *testing.T) {
	t.Parallel()

	grpcClient := &GrpcClient{conf: baseGrpcConf()}

	serviceConfig := grpcClient.createServiceConfig()

	expectedFragments := []string{
		`"loadBalancingPolicy": "round_robin"`,
		`"healthCheckConfig": { "serviceName": "filesync" }`,
		`"name": [{"service": "filesync"}]`,
		`"timeout": "5s"`,
		`"MaxAttempts": 3`,
		`"InitialBackoff": "100ms"`,
		`"MaxBackoff": "2s"`,
		`"BackoffMultiplier": 1.500000`,
		`"RetryableStatusCodes": ["UNAVAILABLE","DEADLINE_EXCEEDED"]`,
	}

	for _, fragment := range expectedFragments {
		if !strings.Contains(serviceConfig, fragment) {
			t.Fatalf("service config does not contain %q\nconfig: %s", fragment, serviceConfig)
		}
	}
}

func TestCreateServiceConfig_WithoutHealthCheck(t *testing.T) {
	t.Parallel()

	conf := baseGrpcConf()
	conf.ShouldUseHealthCheck = false

	grpcClient := &GrpcClient{conf: conf}
	serviceConfig := grpcClient.createServiceConfig()

	if strings.Contains(serviceConfig, "healthCheckConfig") {
		t.Fatalf("healthCheckConfig must not be present when ShouldUseHealthCheck=false: %s", serviceConfig)
	}
}

func TestClientMethods_ReturnClientNotStartedWhenNotStarted(t *testing.T) {
	t.Parallel()

	grpcClient := &GrpcClient{
		conf:      baseGrpcConf(),
		ctx:       context.Background(),
		logger:    testLogger(),
		loggerCtx: context.Background(),
	}

	rootName := mustRootName(t, "root")
	relativePath := mustPath(t, "file.txt")

	if err := grpcClient.DeleteFile(context.Background(), rootName, relativePath); !errors.Is(err, ErrClientNotStarted) {
		t.Fatalf("DeleteFile: expected ErrClientNotStarted, got %v", err)
	}

	if _, err := grpcClient.GetSnapshot(context.Background(), rootName); !errors.Is(err, ErrClientNotStarted) {
		t.Fatalf("GetSnapshot: expected ErrClientNotStarted, got %v", err)
	}

	if _, err := grpcClient.GetFile(context.Background(), rootName, relativePath); !errors.Is(err, ErrClientNotStarted) {
		t.Fatalf("GetFile: expected ErrClientNotStarted, got %v", err)
	}

	if err := grpcClient.PutFile(context.Background(), bytes.NewBufferString("abc"), rootName, relativePath); !errors.Is(err, ErrClientNotStarted) {
		t.Fatalf("PutFile: expected ErrClientNotStarted, got %v", err)
	}

	if err := grpcClient.RenameFile(context.Background(), "uuid", rootName, relativePath, mustPath(t, "new.txt")); !errors.Is(err, ErrClientNotStarted) {
		t.Fatalf("RenameFile: expected ErrClientNotStarted, got %v", err)
	}
}

func TestDeleteFile_UsesExpectedRequestFields(t *testing.T) {
	t.Parallel()

	fakeServiceClient := &fakeFileSyncServiceClient{}
	grpcClient := startedClientWithFakeService(t, fakeServiceClient)

	rootName := mustRootName(t, "root")
	relativePath := mustPath(t, "folder/file.txt")

	err := grpcClient.DeleteFile(context.Background(), rootName, relativePath)
	if err != nil {
		t.Fatalf("DeleteFile returned error: %v", err)
	}

	if fakeServiceClient.deleteFileRequest == nil {
		t.Fatalf("expected DeleteFile request to be captured")
	}

	if fakeServiceClient.deleteFileRequest.RootName != "root" {
		t.Fatalf("unexpected RootName: %q", fakeServiceClient.deleteFileRequest.RootName)
	}

	if fakeServiceClient.deleteFileRequest.RelativePath != "folder/file.txt" {
		t.Fatalf("unexpected RelativePath: %q", fakeServiceClient.deleteFileRequest.RelativePath)
	}
}

func TestRenameFile_UsesExpectedRequestFields(t *testing.T) {
	t.Parallel()

	fakeServiceClient := &fakeFileSyncServiceClient{}
	grpcClient := startedClientWithFakeService(t, fakeServiceClient)

	rootName := mustRootName(t, "root")
	oldRelativePath := mustPath(t, "old.txt")
	newRelativePath := mustPath(t, "new.txt")

	err := grpcClient.RenameFile(context.Background(), "file-uuid", rootName, oldRelativePath, newRelativePath)
	if err != nil {
		t.Fatalf("RenameFile returned error: %v", err)
	}

	if fakeServiceClient.renameFileRequest == nil {
		t.Fatalf("expected RenameFile request to be captured")
	}

	if fakeServiceClient.renameFileRequest.RootName != "root" {
		t.Fatalf("unexpected RootName: %q", fakeServiceClient.renameFileRequest.RootName)
	}

	if fakeServiceClient.renameFileRequest.OldRelativePath != "old.txt" {
		t.Fatalf("unexpected OldRelativePath: %q", fakeServiceClient.renameFileRequest.OldRelativePath)
	}

	if fakeServiceClient.renameFileRequest.NewRelativePath != "new.txt" {
		t.Fatalf("unexpected NewRelativePath: %q", fakeServiceClient.renameFileRequest.NewRelativePath)
	}
}

func TestGetSnapshot_ConvertsProtoSnapshotToDomain(t *testing.T) {
	t.Parallel()

	firstRelativePath, _ := domain.NewPath("b.txt")
	secondRelativePath, _ := domain.NewPath("a.txt")

	fakeServiceClient := &fakeFileSyncServiceClient{
		getSnapshotResponse: &insyncpb.GetSnapshotResponse{
			Snapshot: &insyncpb.Snapshot{
				UnixTime: 123,
				Files: []*insyncpb.FileEntry{
					{
						RelativePath: "b.txt",
						ModifiedUnix: 10,
						SizeBytes:    20,
						IsDirectory:  false,
						Hash:         "hash-b",
						SubtreeSize:  200,
					},
					{
						RelativePath: "a.txt",
						ModifiedUnix: 11,
						SizeBytes:    21,
						IsDirectory:  false,
						Hash:         "hash-a",
						SubtreeSize:  100,
					},
				},
			},
		},
	}

	grpcClient := startedClientWithFakeService(t, fakeServiceClient)

	snapshotWithMetadata, err := grpcClient.GetSnapshot(context.Background(), mustRootName(t, "root"))
	if err != nil {
		t.Fatalf("GetSnapshot returned error: %v", err)
	}

	if len(snapshotWithMetadata.Snapshot.Files) != 2 {
		t.Fatalf("unexpected number of files: %d", len(snapshotWithMetadata.Snapshot.Files))
	}

	if snapshotWithMetadata.Snapshot.Files[0].RelativePath != secondRelativePath {
		t.Fatalf("expected files to be sorted, got first path %q", snapshotWithMetadata.Snapshot.Files[0].RelativePath)
	}

	if snapshotWithMetadata.Snapshot.Files[1].RelativePath != firstRelativePath {
		t.Fatalf("expected files to be sorted, got second path %q", snapshotWithMetadata.Snapshot.Files[1].RelativePath)
	}
}

func TestGetSnapshot_ReturnsConversionError(t *testing.T) {
	t.Parallel()

	fakeServiceClient := &fakeFileSyncServiceClient{
		getSnapshotResponse: &insyncpb.GetSnapshotResponse{
			Snapshot: &insyncpb.Snapshot{
				UnixTime: 123,
				Files: []*insyncpb.FileEntry{
					{
						RelativePath: "file.txt",
						ModifiedUnix: 10,
						SizeBytes:    20,
						IsDirectory:  false,
						Hash:         "",
						SubtreeSize:  100,
					},
				},
			},
		},
	}

	grpcClient := startedClientWithFakeService(t, fakeServiceClient)

	_, err := grpcClient.GetSnapshot(context.Background(), mustRootName(t, "root"))
	if !errors.Is(err, domain.ErrInvalidHash) {
		t.Fatalf("expected ErrInvalidHash, got %v", err)
	}
}

func TestGetFile_ReadsStreamAndReturnsReader(t *testing.T) {
	t.Parallel()

	firstChunk := []byte("hello ")
	secondChunk := []byte("world")

	fakeStream := &fakeGetFileStream{
		responses: []*insyncpb.GetFileResponse{
			{
				Chunk: &insyncpb.FileChunk{
					Data: firstChunk,
				},
			},
			{
				Chunk: &insyncpb.FileChunk{
					Data: secondChunk,
				},
			},
		},
	}

	fakeServiceClient := &fakeFileSyncServiceClient{
		getFileStream: fakeStream,
	}

	grpcClient := startedClientWithFakeService(t, fakeServiceClient)

	reader, err := grpcClient.GetFile(context.Background(), mustRootName(t, "root"), mustPath(t, "folder/file.txt"))
	if err != nil {
		t.Fatalf("GetFile returned error: %v", err)
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}

	if err := reader.Close(); err != nil {
		t.Fatalf("reader.Close returned error: %v", err)
	}

	if string(content) != "hello world" {
		t.Fatalf("unexpected content: %q", string(content))
	}
}

func TestPutFile_SendsInitAndChunkMessages(t *testing.T) {
	t.Parallel()

	fakeStream := &fakePutFileStream{
		closeAndRecvResponse: &emptypb.Empty{},
	}

	fakeServiceClient := &fakeFileSyncServiceClient{
		putFileStream: fakeStream,
	}

	grpcClient := startedClientWithFakeService(t, fakeServiceClient)
	grpcClient.conf.ChunkSizeInBytes = 3

	err := grpcClient.PutFile(
		context.Background(),
		bytes.NewBufferString("abcdefg"),
		mustRootName(t, "root"),
		mustPath(t, "folder/file.txt"),
	)
	if err != nil {
		t.Fatalf("PutFile returned error: %v", err)
	}

	if len(fakeStream.sentMessages) != 4 {
		t.Fatalf("unexpected number of sent messages: %d", len(fakeStream.sentMessages))
	}

	initPayload, ok := fakeStream.sentMessages[0].Payload.(*insyncpb.PutFileRequest_Init)
	if !ok {
		t.Fatalf("first message must be init payload, got %T", fakeStream.sentMessages[0].Payload)
	}

	if initPayload.Init.RootName != "root" {
		t.Fatalf("unexpected init RootName: %q", initPayload.Init.RootName)
	}

	if initPayload.Init.RelativePath != "folder/file.txt" {
		t.Fatalf("unexpected init RelativePath: %q", initPayload.Init.RelativePath)
	}

	expectedChunks := []string{"abc", "def", "g"}
	for chunkIndex, expectedChunk := range expectedChunks {
		chunkPayload, ok := fakeStream.sentMessages[chunkIndex+1].Payload.(*insyncpb.PutFileRequest_Chunk)
		if !ok {
			t.Fatalf("message %d must be chunk payload, got %T", chunkIndex+1, fakeStream.sentMessages[chunkIndex+1].Payload)
		}

		if chunkPayload.Chunk.Index != uint64(chunkIndex) {
			t.Fatalf("unexpected chunk index: got %d, want %d", chunkPayload.Chunk.Index, chunkIndex)
		}

		if string(chunkPayload.Chunk.Data) != expectedChunk {
			t.Fatalf("unexpected chunk data: got %q, want %q", string(chunkPayload.Chunk.Data), expectedChunk)
		}
	}
}

func TestPbFileEntriesToDomain_CollectsValidEntriesAndReturnsErrorForInvalidOnes(t *testing.T) {
	t.Parallel()

	pbEntries := []*insyncpb.FileEntry{
		{
			RelativePath: "valid.txt",
			ModifiedUnix: 10,
			SizeBytes:    20,
			IsDirectory:  false,
			Hash:         "hash-valid",
			SubtreeSize:  100,
		},
		{
			RelativePath: "invalid.txt",
			ModifiedUnix: 11,
			SizeBytes:    21,
			IsDirectory:  false,
			Hash:         "",
			SubtreeSize:  101,
		},
	}

	fileEntries, err := pbFileEntriesToDomain(pbEntries)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if len(fileEntries) != 1 {
		t.Fatalf("unexpected number of converted entries: %d", len(fileEntries))
	}

	if fileEntries[0].RelativePath.String() != "valid.txt" {
		t.Fatalf("unexpected converted entry path: %q", fileEntries[0].RelativePath.String())
	}
}

func TestPbSnapshotToDomain_NilSnapshot(t *testing.T) {
	t.Parallel()

	snapshotWithMetadata, err := pbSnapshotToDomain(nil)
	if err != nil {
		t.Fatalf("pbSnapshotToDomain returned error: %v", err)
	}

	if len(snapshotWithMetadata.Snapshot.Files) != 0 {
		t.Fatalf("unexpected files count: %d", len(snapshotWithMetadata.Snapshot.Files))
	}
}

func TestCreateTransportCreds(t *testing.T) {
	t.Parallel()

	certPath, keyPath, caPath := writeSelfSignedCertFiles(t)

	grpcClient := &GrpcClient{
		conf: &GrpcConf{
			CertPath:   certPath,
			KeyPath:    keyPath,
			CaCertPath: caPath,
			ServerName: "localhost",
		},
		logger:    testLogger(),
		loggerCtx: context.Background(),
	}

	creds, err := grpcClient.createTransportCreds()
	if err != nil {
		t.Fatalf("createTransportCreds returned error: %v", err)
	}

	if creds == nil {
		t.Fatalf("expected non-nil transport credentials")
	}
}

func TestCreateTransportCreds_InvalidCertificatePath(t *testing.T) {
	t.Parallel()

	grpcClient := &GrpcClient{
		conf: &GrpcConf{
			CertPath:   filepath.Join(t.TempDir(), "missing.crt"),
			KeyPath:    filepath.Join(t.TempDir(), "missing.key"),
			CaCertPath: filepath.Join(t.TempDir(), "missing-ca.crt"),
			ServerName: "localhost",
		},
		logger:    testLogger(),
		loggerCtx: context.Background(),
	}

	_, err := grpcClient.createTransportCreds()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func writeSelfSignedCertFiles(t *testing.T) (certPath string, keyPath string, caPath string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	certificateTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certificateDER, err := x509.CreateCertificate(
		rand.Reader,
		&certificateTemplate,
		&certificateTemplate,
		&privateKey.PublicKey,
		privateKey,
	)
	if err != nil {
		t.Fatalf("CreateCertificate: %v", err)
	}

	certificatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certificateDER,
	})

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	tempDir := t.TempDir()

	certPath = filepath.Join(tempDir, "client.crt")
	keyPath = filepath.Join(tempDir, "client.key")
	caPath = filepath.Join(tempDir, "ca.crt")

	if err := os.WriteFile(certPath, certificatePEM, 0o600); err != nil {
		t.Fatalf("WriteFile(cert): %v", err)
	}

	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		t.Fatalf("WriteFile(key): %v", err)
	}

	if err := os.WriteFile(caPath, certificatePEM, 0o600); err != nil {
		t.Fatalf("WriteFile(ca): %v", err)
	}

	return certPath, keyPath, caPath
}

type fakeFileSyncServiceClient struct {
	deleteFileRequest   *insyncpb.DeleteFileRequest
	deleteFileErr       error
	getSnapshotRequest  *insyncpb.GetSnapshotRequest
	getSnapshotResponse *insyncpb.GetSnapshotResponse
	getSnapshotErr      error
	renameFileRequest   *insyncpb.RenameFileRequest
	renameFileErr       error
	getFileStream       insyncpb.FileSyncService_GetFileClient
	getFileErr          error
	putFileStream       insyncpb.FileSyncService_PutFileClient
	putFileErr          error
}

func (f *fakeFileSyncServiceClient) DeleteFile(ctx context.Context, request *insyncpb.DeleteFileRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	f.deleteFileRequest = request
	if f.deleteFileErr != nil {
		return nil, f.deleteFileErr
	}
	return &emptypb.Empty{}, nil
}

func (f *fakeFileSyncServiceClient) GetSnapshot(ctx context.Context, request *insyncpb.GetSnapshotRequest, opts ...grpc.CallOption) (*insyncpb.GetSnapshotResponse, error) {
	f.getSnapshotRequest = request
	if f.getSnapshotErr != nil {
		return nil, f.getSnapshotErr
	}
	return f.getSnapshotResponse, nil
}

func (f *fakeFileSyncServiceClient) GetFile(ctx context.Context, request *insyncpb.GetFileRequest, opts ...grpc.CallOption) (insyncpb.FileSyncService_GetFileClient, error) {
	if f.getFileErr != nil {
		return nil, f.getFileErr
	}
	return f.getFileStream, nil
}

func (f *fakeFileSyncServiceClient) PutFile(ctx context.Context, opts ...grpc.CallOption) (insyncpb.FileSyncService_PutFileClient, error) {
	if f.putFileErr != nil {
		return nil, f.putFileErr
	}
	return f.putFileStream, nil
}

func (f *fakeFileSyncServiceClient) RenameFile(ctx context.Context, request *insyncpb.RenameFileRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	f.renameFileRequest = request
	if f.renameFileErr != nil {
		return nil, f.renameFileErr
	}
	return &emptypb.Empty{}, nil
}

type fakeGetFileStream struct {
	grpc.ClientStream

	responses      []*insyncpb.GetFileResponse
	nextIndex      int
	closeSendCount int
}

func (s *fakeGetFileStream) Recv() (*insyncpb.GetFileResponse, error) {
	if s.nextIndex >= len(s.responses) {
		return nil, io.EOF
	}

	response := s.responses[s.nextIndex]
	s.nextIndex++
	return response, nil
}

func (s *fakeGetFileStream) CloseSend() error {
	s.closeSendCount++
	return nil
}

type fakePutFileStream struct {
	grpc.ClientStream

	sentMessages         []*insyncpb.PutFileRequest
	closeSendCount       int
	closeAndRecvErr      error
	closeAndRecvResponse *emptypb.Empty
}

func (s *fakePutFileStream) Send(request *insyncpb.PutFileRequest) error {
	s.sentMessages = append(s.sentMessages, request)
	return nil
}

func (s *fakePutFileStream) CloseAndRecv() (*emptypb.Empty, error) {
	if s.closeAndRecvResponse == nil {
		s.closeAndRecvResponse = &emptypb.Empty{}
	}
	return s.closeAndRecvResponse, s.closeAndRecvErr
}

func (s *fakePutFileStream) CloseSend() error {
	s.closeSendCount++
	return nil
}
