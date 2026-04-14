package server

import (
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
)

func domainSnapshotToPb(snapshot domain.Snapshot) *insyncpb.Snapshot {
	return &insyncpb.Snapshot{
		RootName: snapshot.RootName.String(),
		UnixTime: snapshot.UnixTime,
		Files:    domainFileEntriesToPb(snapshot.Files),
	}
}
