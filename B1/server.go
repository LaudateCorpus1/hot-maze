package hotmaze

import (
	"net/http"
	"time"

	"cloud.google.com/go/storage"
)

// Server encapsulates the Hot Maze backend.
type Server struct {
	GCPProjectID string

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

// RegisterHandlers registers the handlers
func (s Server) RegisterHandlers() {
	// Static assets: HTML, JS, CSS
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Backend logic
	http.HandleFunc("/secure-urls", s.HandlerGenerateSignedURLs)
	http.HandleFunc("/forget", s.HandlerForgetFile)
}
