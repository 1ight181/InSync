package server

import (
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
)

func domainSnapshotToPb(snapshotWithMetadata domain.SnapshotWithMetadata) *insyncpb.Snapshot {
	return &insyncpb.Snapshot{
		Files:    domainFileEntriesToPb(snapshotWithMetadata.Snapshot.Files),
		UnixTime: snapshotWithMetadata.Metadata.UnixTime,
		RootName: snapshotWithMetadata.Metadata.RootName.String(),
		DeviceId: snapshotWithMetadata.Metadata.DeviceId.String(),
	}
}
