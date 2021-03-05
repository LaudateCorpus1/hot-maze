package hotmaze

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

const (
	validity = 300 * time.Second
)

func (s *Server) HandlerGenerateSignedURLs(w http.ResponseWriter, r *http.Request) {
	if errCORS := s.accessControlAllowHotMaze(w, r); errCORS != nil {
		log.Println(errCORS)
		http.Error(w, errCORS.Error(), http.StatusBadRequest)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "POST only", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	filesize, _ := strconv.Atoi(r.FormValue("filesize"))
	fileUUID, uploadURL, downloadURL, err := s.GenerateURLs(
		ctx,
		r.FormValue("filetype"),
		filesize,
		r.FormValue("filename"),
	)
	if err != nil {
		log.Println("generating signed URLs:", err)
		http.Error(w, "Could not generate signed URLs :(", http.StatusInternalServerError)
		return
	}

	_, err = s.ScheduleForgetFile(ctx, fileUUID)
	if err != nil {
		log.Println("scheduling file expiry:", err)
		// Better fail now, than keeping a user file forever in GCS
		http.Error(w, "Problem with file allocation :(", http.StatusInternalServerError)
		return
	}
	log.Printf("Scheduled deletion for file UUID %q", fileUUID)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"uploadURL":   uploadURL,
		"downloadURL": downloadURL,
	})
}

func (s *Server) GenerateURLs(
	ctx context.Context,
	fileType string,
	fileSize int,
	filename string,

) (fileUUID, uploadURL, downloadURL string, err error) {
	fileUUID = uuid.New().String()
	objectName := "transit/" + fileUUID
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

	// Instead of the full download signed URL which is too big to fit
	// comfortably in a QR-code, we return a short voucher instead.
	downloadURL = s.BackendBaseURL + "/get/" + fileUUID

	return
}

// HandlerUnshortenGetURL redirects a "short" URL to a "long" signed URL.
// Short URL has length ~80.
// Signed download URL has length ~550.
func (s *Server) HandlerUnshortenGetURL(w http.ResponseWriter, r *http.Request) {
	fileUUID := strings.TrimPrefix(r.URL.Path, "/get/")
	objectName := "transit/" + fileUUID
	log.Printf("Redirecting to a new download signed URL for ephemeral resource %q\n", objectName)

	downloadURL, err := storage.SignedURL(
		s.StorageBucket,
		objectName,
		&storage.SignedURLOptions{
			GoogleAccessID: s.StorageAccountID,
			PrivateKey:     s.StoragePrivateKey,
			Method:         "GET",
			Expires:        time.Now().Add(validity),
		})
	if err != nil {
		log.Println("generating download signed URL:", err)
		http.Error(w, "Could not generate download signed URL :(", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, downloadURL, http.StatusFound)
}
