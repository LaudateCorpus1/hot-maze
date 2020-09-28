package hotmaze

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

const (
	validity = 300 * time.Second
)

func (s Server) HandlerUpload(w http.ResponseWriter, r *http.Request) {
	if errCORS := s.accessControlAllowHotMaze(w, r); errCORS != nil {
		log.Println(errCORS)
		http.Error(w, errCORS.Error(), http.StatusBadRequest)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "POST only", http.StatusBadRequest)
		return
	}

	fileData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Could not read file data :(", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	if len(fileData) == 0 {
		http.Error(w, "Empty file :(", http.StatusBadRequest)
		return
	}
	contentType := r.Header.Get("Content-Type")
	log.Println("Received a resource of type", contentType, "and size", len(fileData))

	fileUUID := uuid.New().String()
	log.Println("Saving with UUID", fileUUID)

	ctx := context.Background()
	fsClient, err := firestore.NewClient(ctx, s.GCPProjectID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Problem accessing Firestore :(", http.StatusInternalServerError)
		return
	}
	_, errCreate := fsClient.Doc("C1/"+fileUUID).Create(ctx, map[string]interface{}{
		"data": fileData,
		"type": contentType,
	})
	if errCreate != nil {
		log.Println(errCreate)
		http.Error(w, "Problem writing to Firestore :(", http.StatusInternalServerError)
		return
	}

	_, err = s.ScheduleForgetFile(ctx, fileUUID)
	if err != nil {
		log.Println("scheduling file expiry:", err)
		// Better fail now, than keeping a user file forever in Firestore
		http.Error(w, "Problem with file allocation :(", http.StatusInternalServerError)
		return
	}
	log.Println("File", fileUUID, "is scheduled for deletion")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"downloadURL": s.BackendBaseURL + "/get/" + fileUUID,
	})
}
