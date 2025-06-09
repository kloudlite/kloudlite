package errors

import "errors"

// github.com/kloudlite/api/apps/message-office/internal/domain/platform-edge/repo.go
var ErrEdgeClusterNotAllocated = errors.New("edge cluster not allocated")

// github.com/kloudlite/api/apps/message-office/internal/domain/platform-edge/repo.go
var (
	ErrNoClusterAvailable = errors.New("no cluster available")
	ErrNoClustersInRegion = errors.New("no clusters found in region")
)
