package hotmaze

import "cloud.google.com/go/storage"

// Server encapsulates the Hot Maze backend.
type Server struct {
	StorageClient *storage.Client

	// Service account email e.g. "ephemeral-storage@hot-maze.iam.gserviceaccount.com"
	StorageAccountID string

	// Secret service account private key (PEM).
	// Don't check it in, prefer using Secret Manager.
	StoragePrivateKey []byte
}
