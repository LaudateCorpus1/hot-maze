package hotmaze

import (
	"time"

	"cloud.google.com/go/storage"
)

// Server encapsulates the Hot Maze backend.
type Server struct {
	GCPProjectID string

	BackendBaseURL string

	StorageClient *storage.Client

	// Service account email e.g. "ephemeral-storage@hot-maze.iam.gserviceaccount.com"
	StorageAccountID string

	// Secret service account private key (PEM).
	// Don't check it in, prefer using Secret Manager.
	StoragePrivateKey []byte

	// StorageBucket e.g. "hot-maze.appspot.com"
	StorageBucket string

	StorageFileTTL time.Duration

	CloudTasksQueuePath string
}
