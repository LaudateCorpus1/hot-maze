package hotmaze

import (
	"context"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
)

// All the configuration & initialization happen here, as they
// are necessary to the Cloud Functions exposed below.

const (
	projectID               = "hot-maze"
	backendBaseURL          = "https://us-central1-hot-maze.cloudfunctions.net"
	storageServiceAccountID = "ephemeral-storage@hot-maze.iam.gserviceaccount.com"
	bucket                  = "hot-maze.appspot.com"
	fileDeleteAfter         = 9 * time.Minute
)

var server Server

func init() {
	ctx := context.Background()
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal("Couldn't create Storage client:", err)
	}

	storagePrivateKey, errSecret := AccessSecretVersion("projects/230384242501/secrets/B2-storage-private-key/versions/latest")
	if errSecret != nil {
		log.Fatal("Couldn't read Storage service account private key:", errSecret)
	}

	server = Server{
		GCPProjectID:        projectID,
		BackendBaseURL:      backendBaseURL,
		StorageClient:       storageClient,
		StorageAccountID:    storageServiceAccountID,
		StoragePrivateKey:   storagePrivateKey,
		StorageBucket:       bucket,
		StorageFileTTL:      fileDeleteAfter,
		CloudTasksQueuePath: "projects/hot-maze/locations/us-central1/queues/b2-file-expiry",
	}
}

func B2_SecureURLs(w http.ResponseWriter, r *http.Request) {
	server.HandlerGenerateSignedURLs(w, r)
}

func B2_Get(w http.ResponseWriter, r *http.Request) {
	server.HandlerUnshortenGetURL(w, r)
}

func B2_Forget(w http.ResponseWriter, r *http.Request) {
	server.HandlerForgetFile(w, r)
}
