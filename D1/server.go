package hotmaze

import (
	"net/http"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/firestore"
)

// Server encapsulates the Hot Maze backend.
type Server struct {
	GCPProjectID string

	BackendBaseURL string

	FirestoreClient *firestore.Client

	TasksClient *cloudtasks.Client

	StorageFileTTL time.Duration

	CloudTasksQueuePath string
}

// RegisterHandlers registers the handlers
func (s Server) RegisterHandlers() {
	// Static assets: HTML, JS, CSS
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})
	http.HandleFunc("/terms.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/terms.html")
	})
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Backend logic
	http.HandleFunc("/upload", s.HandlerUpload)
	http.HandleFunc("/get/", s.HandlerDownload)
	http.HandleFunc("/forget", s.HandlerForgetFile)
}
