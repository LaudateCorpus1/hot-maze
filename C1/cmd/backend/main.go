package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	hotmaze "github.com/Deleplace/hot-maze/C1"
)

const (
	projectID = "hot-maze"
	// "c1" is the version of this App Engine app
	// "uc.r" is the app location (regional route)
	backendBaseURL          = "https://c1-dot-hot-maze.uc.r.appspot.com"
	storageServiceAccountID = "ephemeral-storage@hot-maze.iam.gserviceaccount.com"
	bucket                  = "hot-maze.appspot.com"
	fileDeleteAfter         = 9 * time.Minute
)

func main() {
	ctx := context.Background()
	fsClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalln("Problem accessing Firestore :(")
	}
	defer fsClient.Close()

	server := hotmaze.Server{
		GCPProjectID:        projectID,
		BackendBaseURL:      backendBaseURL,
		FirestoreClient:     fsClient,
		StorageFileTTL:      fileDeleteAfter,
		CloudTasksQueuePath: "projects/hot-maze/locations/us-central1/queues/c1-file-expiry",
	}
	server.RegisterHandlers()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	log.Fatal(err)
}
