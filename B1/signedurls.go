package hotmaze

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

const (
	validity = 300 * time.Second
)

func (s Server) HandlerGenerateSignedURLs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "localhost:*")
	if errCORS := s.accessControlAllowHotMaze(w, r); errCORS != nil {
		log.Println(errCORS)
		fmt.Fprint(w, errCORS)
		return
	}
	ctx := context.Background()
	filesize, _ := strconv.Atoi(r.FormValue("filesize"))
	uploadURL, downloadURL, err := s.GenerateURLs(
		ctx,
		r.FormValue("filetype"),
		filesize,
		r.FormValue("filename"),
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Could not generate signed URLs :(")
		log.Println("generating signed URLs:", err)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"uploadURL":   uploadURL,
		"downloadURL": downloadURL,
	})
}

func (s Server) GenerateURLs(
	ctx context.Context,
	fileType string,
	fileSize int,
	filename string,

) (uploadURL, downloadURL string, err error) {
	objectName := "transit/" + uuid.New().String()
	log.Printf("Creating URLs for ephemeral resource %q\n", objectName)

	uploadURL, err = storage.SignedURL(
		s.StorageBucket,
		objectName,
		&storage.SignedURLOptions{
			GoogleAccessID: s.StorageAccountID,
			PrivateKey:     s.StoragePrivateKey,
			Method:         "PUT",
			Expires:        time.Now().Add(validity),
			ContentType:    fileType,
		})
	if err != nil {
		return
	}

	downloadURL, err = storage.SignedURL(
		s.StorageBucket,
		objectName,
		&storage.SignedURLOptions{
			GoogleAccessID: s.StorageAccountID,
			PrivateKey:     s.StoragePrivateKey,
			Method:         "GET",
			Expires:        time.Now().Add(validity),
		})

	return
}
