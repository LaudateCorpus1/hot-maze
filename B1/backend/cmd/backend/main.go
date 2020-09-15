package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
	hotmaze "github.com/Deleplace/hot-maze/B1/backend"
	"golang.org/x/net/context"
)

const (
	projectId               = "hot-maze"
	storageServiceAccountID = "ephemeral-storage@hot-maze.iam.gserviceaccount.com"
)

func main() {
	ctx := context.Background()
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal("Couldn't create Storage client:", err)
	}

	storagePrivateKey, errSecret := hotmaze.AccessSecretVersion("projects/230384242501/secrets/B1-storage-private-key/versions/latest")
	if errSecret != nil {
		log.Fatal("Couldn't read Storage service account private key:", err)
	}

	// storagePrivateKey, err := ioutil.ReadFile("/tmp/hot-maze-9f2d6311662c.pem")
	// if err != nil {
	// 	log.Fatal("Couldn't read Storage service account private key:", err)
	// }

	//
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../../../web/"))))

	server := hotmaze.Server{
		StorageClient:     storageClient,
		StorageAccountID:  storageServiceAccountID,
		StoragePrivateKey: storagePrivateKey,
	}
	http.HandleFunc("/secure-urls", server.HandlerGenerateSignedURLs)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	log.Fatal(err)
}
